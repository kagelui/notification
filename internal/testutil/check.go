package testutil

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// Ok tests if err is nil
func Ok(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// Equals fails the test if exp is not equal to act.
func Equals(tb testing.TB, exp, act interface{}, opts ...cmp.Option) {
	tb.Helper()
	if !cmp.Equal(exp, act, opts...) {
		_, file, line, _ := runtime.Caller(1)
		tb.Fatalf("\033[31m%s:%d:\n\n%v\033[39m\n\n", filepath.Base(file), line, cmp.Diff(exp, act, opts...))
	}
}

// Asserts fails the test if the condition is false.
func Asserts(tb testing.TB, condition bool, msg string, v ...interface{}) {
	tb.Helper()
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		tb.Fatalf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
	}
}

// CompareError tests if the err contains probe
func CompareError(t *testing.T, probe string, err error) {
	t.Helper()
	if err == nil {
		if probe != "" {
			t.Errorf("err is nil but expect it to contain %v", probe)
		}
	} else if probe == "" {
		t.Errorf("err is %v but expect nil", err.Error())
	} else if !strings.Contains(err.Error(), probe) {
		t.Errorf("err is %v, expected it to contain %v", err.Error(), probe)
	}
}

// MustOpen will panic if error
func MustOpen(fileName string) *os.File {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	return file
}

// CheckTimeApproximately fails the test t if two times are not close enough
func CheckTimeApproximately(t *testing.T, expectedTime, actualTime time.Time) {
	if !TimesAreCloseEnough(expectedTime, actualTime) {
		t.Errorf("expected time %v and actual time %v is not close enough", expectedTime, actualTime)
	}
}

// TimesAreCloseEnough tells if two times are within one minute
func TimesAreCloseEnough(a, b time.Time) bool {
	const limit = time.Minute
	if a.After(b) {
		return a.Sub(b) < limit
	}
	return b.Sub(a) < limit
}

type roundTripFunc func(req *http.Request) *http.Response

// RoundTrip implements RoundTripper
func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}
