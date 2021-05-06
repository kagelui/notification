package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kagelui/notification/internal/pkg/web"
)

type messageStore interface {
	InsertCallbackThenDo(ctx context.Context, productID, productType, payload, businessID string, timeout time.Duration) error
}

type callbackRequest struct {
	ProductID   string `json:"product_id"`
	ProductType string `json:"product_type"`
	Payload     string `json:"payload"`
	BusinessID  string `json:"business_id"`
}

// StoreCallbackThenSend stores the callback and perform the call back
func StoreCallbackThenSend(store messageStore, timeout time.Duration) web.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		req := callbackRequest{}
		if er := json.NewDecoder(r.Body).Decode(&req); er != nil {
			return errParsingRequest
		}
		if err := store.InsertCallbackThenDo(r.Context(), req.ProductID, req.ProductType, req.Payload, req.BusinessID, timeout); err != nil {
			return web.NewError(err, "error performing callback")
		}
		web.RespondJSON(r.Context(), w, "ok", nil)
		return nil
	}
}
