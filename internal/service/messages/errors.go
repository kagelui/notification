package messages

import (
	"net/http"

	"github.com/kagelui/notification/internal/pkg/web"
)

// ErrMerchantInfoNotLoaded occurs when the merchant relation is not loaded
var ErrMerchantInfoNotLoaded = web.Error{Status: http.StatusInternalServerError, Code: "info_not_loaded", Desc: "merchant_info_empty"}
