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

type ConfirmationResult struct {
	Order      Order                 // The confirmed order.
	Membership reader.Membership     // The updated membership.
	Snapshot   reader.MemberSnapshot // // Snapshot of previous membership
}
