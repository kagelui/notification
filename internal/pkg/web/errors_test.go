package web

import (
	"errors"
	"net/http"
	"testing"

	"github.com/kagelui/notification/internal/testutil"
)

func TestWithStack(t *testing.T) {
	testCases := []struct {
		given          error
		expectedStatus int
		expectedCode   string
		expectedDesc   string
		expectedErrMsg string
	}{
		{&Error{Status: http.StatusBadRequest, Code: "web_error", Desc: "web error"}, http.StatusBadRequest, "web_error", "web error", "web error"},
		{&Error{Status: http.StatusBadRequest, Code: "web_error", Desc: "web error", Err: errors.New("something else")}, http.StatusBadRequest, "web_error", "web error", "something else"},
		{errors.New("error"), http.StatusInternalServerError, "internal_error", "error", "error"},
	}

	for _, tc := range testCases {
		t.Run(tc.given.Error(), func(t *testing.T) {
			result := WithStack(tc.given)

			testutil.Asserts(t, result.Err != nil, "err should not be nil")
			testutil.Equals(t, tc.expectedStatus, result.Status)
			testutil.Equals(t, tc.expectedCode, result.Code)
			testutil.Equals(t, tc.expectedDesc, result.Desc)
			testutil.Equals(t, tc.expectedErrMsg, result.Err.Error())
			testutil.Equals(t, tc.given.Error(), result.Desc)
		})
	}
}

func TestNewError(t *testing.T) {
	testCases := []struct {
		given          error
		message        string
		expectedStatus int
		expectedCode   string
	}{
		{&Error{Status: http.StatusBadRequest, Code: "web_error", Desc: "web error"}, "custom message", http.StatusBadRequest, "web_error"},
		{errors.New("error"), "custom message", http.StatusInternalServerError, "internal_error"},
	}

	for _, tc := range testCases {
		t.Run(tc.given.Error(), func(t *testing.T) {
			result := NewError(tc.given, tc.message)

			testutil.Asserts(t, result.Err != nil, "err should not be nil")
			testutil.Equals(t, tc.expectedStatus, result.Status)
			testutil.Equals(t, tc.expectedCode, result.Code)
			testutil.Equals(t, tc.message, result.Desc)
		})
	}
}
