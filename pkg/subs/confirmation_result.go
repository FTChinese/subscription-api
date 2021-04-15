package subs

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
)

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
	Payment    PaymentResult         `json:"payment"` // Empty if order is already confirmed.
	Order      Order                 `json:"order"`   // The confirmed order
	Invoices   Invoices              `json:"-"`
	Membership reader.Membership     `json:"membership"` // The updated membership. Empty if order is already confirmed.
	Snapshot   reader.MemberSnapshot `json:"-"`
	Notify     bool                  `json:"-"`
}

func NewConfirmationResult(p ConfirmationParams) (ConfirmationResult, error) {

	invoices, err := p.invoices()
	if err != nil {
		return ConfirmationResult{}, err
	}

	newM, err := invoices.membership(p.Order.UserIDs, p.Member)
	if err != nil {
		return ConfirmationResult{}, err
	}

	return ConfirmationResult{
		Payment:    p.Payment,
		Order:      p.Order.Confirmed(p.Payment.ConfirmedUTC, invoices.Purchased.DateTimePeriod),
		Invoices:   invoices,
		Membership: newM,
		Snapshot:   p.snapshot(),
		Notify:     true,
	}, nil
}
