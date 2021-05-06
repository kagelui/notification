package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kagelui/notification/internal/pkg/web"
	"github.com/kagelui/notification/internal/testutil"
)

func TestWrapError(t *testing.T) {
	type args struct {
		h web.HandlerFunc
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantBody   string
		wantHeader string
	}{
		{
			name: "error",
			args: args{h: func(w http.ResponseWriter, r *http.Request) error {
				return &web.Error{Status: http.StatusBadRequest, Code: "code", Desc: "desc"}
			}},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"code","error_description":"desc"}`,
			wantHeader: "application/json",
		},
		{
			name: "no error",
			args: args{h: func(w http.ResponseWriter, r *http.Request) error {
				return nil
			}},
			wantStatus: http.StatusOK,
			wantBody:   "",
			wantHeader: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/some/url", nil)

			web.Handler{H: tt.args.h}.ServeHTTP(rr, req)

			testutil.Equals(t, tt.wantStatus, rr.Code)
			testutil.Equals(t, tt.wantBody, rr.Body.String())
			testutil.Equals(t, tt.wantHeader, rr.Header().Get("Content-Type"))
		})
	}
}
