package messages

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kagelui/notification/internal/models/bmodels"
	"github.com/kagelui/notification/internal/testutil"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

func TestCallbackClient_doOneCallback(t *testing.T) {
	errorMsg := "mock error"
	errorResp := &http.Response{
		Status:        "500 Internal Error",
		StatusCode:    http.StatusInternalServerError,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(errorMsg)),
		ContentLength: int64(len(errorMsg)),
		Request:       &http.Request{},
		Header:        make(http.Header, 0),
	}

	type fields struct {
		Client *http.Client
	}
	type args struct {
		url     string
		token   string
		payload string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr string
	}{
		{
			name: "fail",
			fields: fields{
				Client: testutil.NewTestClient(func(req *http.Request) *http.Response {
					data, err := ioutil.ReadAll(req.Body)
					if err != nil {
						t.Errorf(err.Error())
					}
					if !strings.Contains(req.URL.String(), "failure_url") ||
						req.Header.Get(tokenHeaderKey) != "some token" || string(data) != "{}" {
						return &http.Response{StatusCode: http.StatusOK}
					}
					return errorResp
				}),
			},
			args: args{
				url:     "failure_url",
				token:   "some token",
				payload: "{}",
			},
			wantErr: "callback error: 500, response: mock error",
		},
		{
			name: "success",
			fields: fields{
				Client: testutil.NewTestClient(func(req *http.Request) *http.Response {
					data, err := ioutil.ReadAll(req.Body)
					if err != nil {
						t.Errorf(err.Error())
					}
					if !strings.Contains(req.URL.String(), "success") ||
						req.Header.Get(tokenHeaderKey) != "some token" || string(data) != "{}" {
						return errorResp
					}
					return &http.Response{StatusCode: http.StatusOK}
				}),
			},
			args: args{
				url:     "success",
				token:   "some token",
				payload: "{}",
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CallbackClient{
				Client: tt.fields.Client,
			}
			testutil.CompareError(t, tt.wantErr, c.doOneCallback(tt.args.url, tt.args.token, tt.args.payload))
		})
	}
}

func TestCallbackClient_DoCallback(t *testing.T) {
	ctx := context.TODO()
	errorMsg := "mock error"
	errorResp := &http.Response{
		Status:        "500 Internal Error",
		StatusCode:    http.StatusInternalServerError,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(errorMsg)),
		ContentLength: int64(len(errorMsg)),
		Request:       &http.Request{},
		Header:        make(http.Header, 0),
	}

	payloadJSON := types.JSON{}
	testutil.Ok(t, payloadJSON.Marshal([]byte("")))

	type fixture struct {
		merchant bmodels.Merchant
		url      bmodels.CallbackURL
		message  bmodels.Message
	}
	type fields struct {
		Client *http.Client
	}
	tests := []struct {
		name    string
		respErr bool
		f       fixture
		fields  fields
		wantErr string
	}{
		{
			name:    "success",
			respErr: false,
			f: fixture{
				merchant: bmodels.Merchant{
					ID:         92137,
					BusinessID: "merchant0",
					Token:      "some token",
				},
				url: bmodels.CallbackURL{
					ID:          32916,
					BusinessID:  "merchant0",
					ProductID:   "va",
					CallbackURL: "success_url",
				},
				message: bmodels.Message{
					ID:               uuid.New().String(),
					ProductID:        "va",
					ProductType:      "something",
					Payload:          payloadJSON,
					MerchantID:       92137,
					RetryCount:       0,
					NextDeliveryTime: time.Now(),
					Status:           MessageDeliveryStatusPending,
				},
			},
			fields: fields{
				Client: testutil.NewTestClient(func(req *http.Request) *http.Response {
					data, err := ioutil.ReadAll(req.Body)
					if err != nil {
						t.Errorf(err.Error())
					}
					if !strings.Contains(req.URL.String(), "success_url") ||
						req.Header.Get(tokenHeaderKey) != "some token" || string(data) != `""` {
						return errorResp
					}
					return &http.Response{StatusCode: http.StatusOK}
				}),
			},
			wantErr: "",
		},
		{
			name:    "response error",
			respErr: true,
			f: fixture{
				merchant: bmodels.Merchant{
					ID:         92137,
					BusinessID: "merchant0",
					Token:      "some token",
				},
				url: bmodels.CallbackURL{
					ID:          32916,
					BusinessID:  "merchant0",
					ProductID:   "va",
					CallbackURL: "success_url",
				},
				message: bmodels.Message{
					ID:               uuid.New().String(),
					ProductID:        "va",
					ProductType:      "something",
					Payload:          payloadJSON,
					MerchantID:       92137,
					RetryCount:       0,
					NextDeliveryTime: time.Now(),
					Status:           MessageDeliveryStatusPending,
				},
			},
			fields: fields{
				Client: testutil.NewTestClient(func(req *http.Request) *http.Response {
					data, err := ioutil.ReadAll(req.Body)
					if err != nil {
						t.Errorf(err.Error())
					}
					if !strings.Contains(req.URL.String(), "success_url") ||
						req.Header.Get(tokenHeaderKey) != "some token" || string(data) != `""` {
						return &http.Response{StatusCode: http.StatusOK}
					}
					return errorResp
				}),
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.MustBegin()
			testutil.Ok(t, tt.f.merchant.Insert(ctx, tx, boil.Infer()))
			testutil.Ok(t, tt.f.url.Insert(ctx, tx, boil.Infer()))
			testutil.Ok(t, tt.f.message.Insert(ctx, tx, boil.Infer()))

			tt.f.message.R = bmodels.Message{}.R.NewStruct()
			tt.f.message.R.Merchant = &tt.f.merchant
			currRetryTime, currRetryCount := tt.f.message.NextDeliveryTime, tt.f.message.RetryCount

			c := CallbackClient{
				Client: tt.fields.Client,
			}
			err := c.DoCallback(ctx, tx, &tt.f.message)
			testutil.CompareError(t, tt.wantErr, err)

			if err == nil {
				testutil.Ok(t, tt.f.message.Reload(ctx, tx))
				if tt.respErr {
					testutil.Equals(t, MessageDeliveryStatusFailed, tt.f.message.Status)
					testutil.Equals(t, currRetryCount+1, tt.f.message.RetryCount)
					testutil.CheckTimeApproximately(t, getRetryTime(currRetryTime, currRetryCount), tt.f.message.NextDeliveryTime)
				} else {
					testutil.Equals(t, MessageDeliveryStatusSuccess, tt.f.message.Status)
				}
			}

			testutil.Ok(t, tx.Rollback())
		})
	}
}
