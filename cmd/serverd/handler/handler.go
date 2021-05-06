package handler

import (
	"net/http"

	"github.com/kagelui/notification/internal/pkg/web"
)

// WrapError wraps the web.HandlerFunc to standard http.HandlerFunc with error handling
func WrapError(h web.HandlerFunc) http.HandlerFunc {
	// wraps error reporter
	h = web.Wrap(h, web.ReportErrorWrapper())

	// Web handler
	wh := web.Handler{H: h}

	return wh.ServeHTTP
}

