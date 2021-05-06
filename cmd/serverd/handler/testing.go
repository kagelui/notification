package handler

import (
	"context"
	"testing"
	"time"

	"github.com/kagelui/notification/internal/testutil"
)

type mockMessageStore struct {
	T *testing.T
	Ms messageIOSuite
}

type messageIOSuite struct {
	ProductID string
	ProductType string
	Payload string
	BusinessID string
	Timeout time.Duration
	Err error
}

func (s mockMessageStore) InsertCallbackThenDo(_ context.Context, productID, productType, payload, businessID string, timeout time.Duration) error {
	testutil.Equals(s.T, s.Ms.ProductID, productID)
	testutil.Equals(s.T, s.Ms.ProductType, productType)
	testutil.Equals(s.T, s.Ms.Payload, payload)
	testutil.Equals(s.T, s.Ms.BusinessID, businessID)
	testutil.Equals(s.T, s.Ms.Timeout, timeout)
	return s.Ms.Err
}
