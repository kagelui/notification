package web

import (
	"net/http"

	"github.com/kagelui/notification/internal/pkg/loglib"
)

// ReportError handles reporting of standard and Error errors
func ReportError(err error, r *http.Request) {
	if err == nil {
		return
	}

	logger := loglib.GetLogger(r.Context())

	logger.ErrorF(err.Error())
	// additional error reporting if necessary
}

// ReportErrorWrapper provides wrapper implementation to HandlerFunc for reporting Error
func ReportErrorWrapper() HandlerWrapper {
	return func(next HandlerFunc) HandlerFunc {
		fn := func(w http.ResponseWriter, r *http.Request) error {
			err := next(w, r)

			// Report error
			if err != nil {
				ReportError(err, r)
			}

			return err
		}
		return fn
	}
}
