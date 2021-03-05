package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/invoice"
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
	Payment    PaymentResult         `json:"payment"`    // Empty if order is already confirmed.
	Order      Order                 `json:"order"`      // The confirmed order
	Invoices   []invoice.Invoice     `json:"-"`          // Each order creates an invoice; for upgrading, there's an additional carry-over invoice.
	Membership reader.Membership     `json:"membership"` // The updated membership. Empty if order is already confirmed.
	Snapshot   reader.MemberSnapshot `json:"-"`
	Notify     bool                  `json:"-"`
}

func NewConfirmationResult(p ConfirmationParams) (ConfirmationResult, error) {

	cfmInv, err := p.invoice()
	if err != nil {
		return ConfirmationResult{}, err
	}

	invoices := []invoice.Invoice{
		cfmInv,
	}

	if p.Order.Kind == enum.OrderKindUpgrade {
		invoices = append(
			invoices,
			invoice.NewFromCarryOver(p.Member, addon.SourceUpgradeCarryOver).
				WithOrderID(p.Order.ID),
		)
	}

	return ConfirmationResult{
		Payment:    p.Payment,
		Order:      p.confirmedOrder(cfmInv.DateTimePeriod),
		Invoices:   invoices,
		Membership: p.membership(cfmInv),
		Snapshot:   p.snapshot(),
		Notify:     true,
	}, nil
}

// MustNewConfirmationResult is the panic version of NewConfirmationResult
func MustNewConfirmationResult(p ConfirmationParams) ConfirmationResult {
	result, err := NewConfirmationResult(p)
	if err != nil {
		panic(err)
	}
	return result
}
