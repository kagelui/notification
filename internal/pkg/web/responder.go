package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kagelui/notification/internal/pkg/loglib"
)

// GenericErrorMessage is the human friendly error message for HTTP 5XX
const GenericErrorMessage = "Sorry, there was a problem. Please try again later."

// RespondJSON writes JSON as http response
func RespondJSON(ctx context.Context, w http.ResponseWriter, object interface{}, headers map[string]string) {
	logger := loglib.GetLogger(ctx)

	// Handle json marshalling error
	respBytes, err := json.Marshal(object)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.ErrorF("[Web responder] JSON marshal error: %s", err)
		return
	}

	// Set HTTP headers
	w.Header().Set("Content-Type", "application/json")
	if headers != nil {
		for key, value := range headers {
			w.Header().Set(key, value)
		}
	}

	// Handle web error
	status := http.StatusOK
	switch werr := object.(type) {
	case *Error:
		// Log raw error response
		logger.ErrorF("[Web responder] Web error: %d %s %s", werr.Status, werr.Code, werr.Desc)

		// 5XX (except 503) should be sanitized before showing to human
		if werr.Status >= 500 && werr.Status != http.StatusServiceUnavailable {
			werr.Desc = GenericErrorMessage
			respBytes, _ = json.Marshal(werr)
		}

		// Add logger fields
		logger = logger.WithField("error", "true")
		status = werr.Status
	}

	// Log response body
	logger.WithField("status", status).
		InfoF("[Web responder] Wrote %d bytes: %v", len(respBytes), string(respBytes))

	// Write response
	w.WriteHeader(status)
	w.Write(respBytes)
}
