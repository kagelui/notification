package messages

import (
	"context"
	"time"

	"github.com/kagelui/notification/internal/models/bmodels"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const maxRetry = 6

// RetrieveAllRetryMessages returns all messages that should be retried
func RetrieveAllRetryMessages(ctx context.Context, db Inquirer) ([]*bmodels.Message, error) {
	return bmodels.Messages(
		bmodels.MessageWhere.Status.EQ(MessageDeliveryStatusFailed),
		bmodels.MessageWhere.RetryCount.LT(maxRetry),
		bmodels.MessageWhere.NextDeliveryTime.LT(time.Now()),
		qm.Load(bmodels.MessageRels.Merchant),
	).All(ctx, db)
}
