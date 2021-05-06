package messages

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/kagelui/notification/internal/models/bmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// MarkMessagesPending marks all message pending delivery
func MarkMessagesPending(ctx context.Context, db *sqlx.DB, slice bmodels.MessageSlice) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	for _, m := range slice {
		m.Status = MessageDeliveryStatusPending
		if _, err = m.Update(ctx, db, boil.Whitelist(bmodels.MessageColumns.Status, bmodels.MerchantColumns.UpdatedAt)); err != nil {
			return err
		}
	}
	return nil
}
