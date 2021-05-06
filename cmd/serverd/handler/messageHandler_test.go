package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kagelui/notification/internal/pkg/web"
	"github.com/kagelui/notification/internal/testutil"
)

func TestStoreCallbackThenSend(t *testing.T) {
	type args struct {
		store   messageStore
		timeout time.Duration
	}
	tests := []struct {
		name         string
		args         args
		request      interface{}
		expectedCode int
		expectedBody string
	}{
		{
			name:         "bad request",
			args:         args{
				store: mockMessageStore{
					T:  t,
					Ms: messageIOSuite{
						ProductID:   "abc",
						ProductType: "efg",
						Payload:     "{}",
						BusinessID:  "user00",
						Timeout:     time.Minute,
						Err:         nil,
					},
				},
				timeout: time.Minute,
			},
			request:      "random string",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"400 Bad Request","error_description":"cannot parse request"}`,
		},
		{
			name:         "naughty store",
			args:         args{
				store: mockMessageStore{
					T:  t,
					Ms: messageIOSuite{
						ProductID:   "abc",
						ProductType: "efg",
						Payload:     "{}",
						BusinessID:  "user00",
						Timeout:     time.Minute,
						Err:         fmt.Errorf("mock error"),
					},
				},
				timeout: time.Minute,
			},
			request:      callbackRequest{
				ProductID:   "abc",
				ProductType: "efg",
				Payload:     "{}",
				BusinessID:  "user00",
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"internal_error","error_description":"Sorry, there was a problem. Please try again later."}`,
		},
		{
			name:         "all good",
			args:         args{
				store: mockMessageStore{
					T:  t,
					Ms: messageIOSuite{
						ProductID:   "abc",
						ProductType: "efg",
						Payload:     "{}",
						BusinessID:  "user00",
						Timeout:     time.Minute,
						Err:         nil,
					},
				},
				timeout: time.Minute,
			},
			request:      callbackRequest{
				ProductID:   "abc",
				ProductType: "efg",
				Payload:     "{}",
				BusinessID:  "user00",
			},
			expectedCode: http.StatusOK,
			expectedBody: `"ok"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, err := json.Marshal(tt.request)
			testutil.Ok(t, err)
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
			rr := httptest.NewRecorder()
			web.Handler{H: StoreCallbackThenSend(tt.args.store, tt.args.timeout)}.ServeHTTP(rr, req)
			testutil.Equals(t, tt.expectedCode, rr.Code)
			testutil.Equals(t, tt.expectedBody, rr.Body.String())
		})
	}
}
