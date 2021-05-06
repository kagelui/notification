package loglib

import (
	"context"
	"testing"

	"github.com/kagelui/notification/internal/testutil"
)

func TestContext(t *testing.T) {
	ctx := context.Background()
	result := GetLogger(ctx)
	testutil.Asserts(t, result != nil, "should have logger when not set")

	given := NewLogger(nil)
	ctx = SetLogger(ctx, given)
	resultNow, ok := HasLogger(ctx)
	testutil.Asserts(t, ok, "should have logger")
	testutil.Equals(t, given, resultNow)

	result = GetLogger(ctx)
	testutil.Equals(t, given, resultNow)
}
