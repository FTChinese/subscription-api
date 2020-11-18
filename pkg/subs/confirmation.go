package subs

import "github.com/FTChinese/subscription-api/pkg/reader"

const StmtSaveConfirmResult = `
INSERT INTO premium.confirmation_result
SET order_id = :order_id,
	failed = :failed,
	created_utc = UTC_TIMESTAMP()`

type ConfirmError struct {
	OrderID string `db:"order_id"`
	Message string `db:"failed"`
	Retry   bool
}

func (c ConfirmError) Error() string {
	return c.Message
}

// ConfirmationResult contains all the data in the process of confirming an order.
// This is also used as the http response for manual confirmation.
type ConfirmationResult struct {
	orderModified bool
	Order         Order `json:"order"` // The confirmed order.
	memberUpdated bool
	Membership    reader.Membership     `json:"membership"` // The updated membership. Empty if order is already confirmed.
	Payment       PaymentResult         `json:"payment"`    // Empty if order is already confirmed.
	Snapshot      reader.MemberSnapshot `json:"-"`          // Snapshot of previous membership
}

// Modified checks if Order and Membership are actually modified.
// Usually if an order is confirmed, membership will be modified.
// There might be case while an order is confirmed, the membership is not changed. We treat it as out of sync.
// For verification, an order should already confirmed and membership updated. In such case, nothing modified and data should be sent to client as is.
// As long as it returns true, membership must be modified..
func (r *ConfirmationResult) Modified() bool {
	// As long as order is modified, all data are changed.
	if r.orderModified {
		return true
	}

	// If order is not modified, membership is modified if it is out of sync; otherwise it is not modified.
	return r.memberUpdated
}

// ShouldUpdateOrder checks if the order is modified.
func (r *ConfirmationResult) ShouldUpdateOrder() bool {
	return r.orderModified
}
