package messages

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kagelui/notification/internal/models/bmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const tokenHeaderKey = "x-callback-token"

type CallbackClient struct {
	Client     *http.Client
}

// Inquirer unifies *sql.DB and *sql.Tx, facilitating unit tests
type Inquirer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// DoCallback carries out the callback, note that messageWithMerchantInfo must contain the merchant info
func (c CallbackClient) DoCallback(ctx context.Context, db Inquirer, messageWithMerchantInfo bmodels.Message) error {
	if messageWithMerchantInfo.R == nil || messageWithMerchantInfo.R.Merchant == nil {
		return ErrMerchantInfoNotLoaded
	}

	urlRecord, err := bmodels.CallbackUrls(
		bmodels.CallbackURLWhere.ProductID.EQ(messageWithMerchantInfo.ProductID),
		bmodels.CallbackURLWhere.BusinessID.EQ(messageWithMerchantInfo.R.Merchant.BusinessID)).One(ctx, db)
	if err != nil {
		return err
	}

	if c.doOneCallback(urlRecord.CallbackURL, messageWithMerchantInfo.R.Merchant.Token, messageWithMerchantInfo.Payload.String()) != nil {
		messageWithMerchantInfo.Status = MessageDeliveryStatusFailed
		messageWithMerchantInfo.NextDeliveryTime = getRetryTime(messageWithMerchantInfo.NextDeliveryTime, messageWithMerchantInfo.RetryCount)
		messageWithMerchantInfo.RetryCount++
		_, e := messageWithMerchantInfo.Update(ctx, db,
			boil.Whitelist(bmodels.MessageColumns.Status,
				bmodels.MessageColumns.NextDeliveryTime,
				bmodels.MessageColumns.RetryCount, bmodels.MessageColumns.UpdatedAt))
		return e
	}

	messageWithMerchantInfo.Status = MessageDeliveryStatusSuccess
	_, e := messageWithMerchantInfo.Update(ctx, db, boil.Whitelist(bmodels.MessageColumns.Status, bmodels.MessageColumns.UpdatedAt))
	return e
}

func (c CallbackClient) doOneCallback(url, token, payload string) error {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}
	req.Header.Set(tokenHeaderKey, token)
	resp, e := c.Client.Do(req)
	if e != nil {
		return e
	}
	if resp.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return fmt.Errorf("callback error: %v, response: %v", resp.StatusCode, string(data))
	}
	return nil
}
