package messages

import (
	"context"
	"github.com/google/uuid"
	"net/http"
	"time"

	"github.com/kagelui/notification/internal/models/bmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

// ModelStore contains a reference to the DB connection and provides the service to handlers
type ModelStore struct {
	DB Inquirer
}

func (m ModelStore) InsertCallbackThenDo(ctx context.Context, productID, productType, payload, businessID string, timeout time.Duration) error {
	// insert the callback
	merchant, err := bmodels.Merchants(bmodels.MerchantWhere.BusinessID.EQ(businessID)).One(ctx, m.DB)
	if err != nil {
		return err
	}

	payloadJSON := types.JSON{}
	if err = payloadJSON.Marshal(payload); err != nil {
		return err
	}

	message := bmodels.Message{
		ID: uuid.New().String(),
		ProductID:        productID,
		ProductType:      productType,
		Payload:          payloadJSON,
		MerchantID:       merchant.ID,
		RetryCount:       0,
		Status:           MessageDeliveryStatusPending,
	}
	if err = message.Insert(ctx, m.DB, boil.Infer()); err != nil {
		return err
	}

	message.R = message.R.NewStruct()
	message.R.Merchant = merchant

	// perform callback
	go func() {
		httpClient := http.DefaultClient
		httpClient.Timeout = timeout
		client := CallbackClient{Client: httpClient}

		client.DoCallback(context.Background(), m.DB, &message)
	}()

	return nil
}
