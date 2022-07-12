package ftcpay

import (
	"github.com/FTChinese/go-rest/enum"
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
	Payment    PaymentResult              `json:"payment"` // Empty if order is already confirmed.
	Order      Order                      `json:"order"`   // The confirmed order
	Invoices   Invoices                   `json:"-"`
	Membership reader.Membership          `json:"membership"` // The updated membership. Empty if order is already confirmed.
	Versioned  reader.MembershipVersioned `json:"-"`
	Notify     bool                       `json:"-"`
}

// NewConfirmationResult confirms an order based on the payment result and
// current membership. The order will be updated, invoices will be created,
// membership will be updated and a snapshot will be taken.
func NewConfirmationResult(p ConfirmationParams) (ConfirmationResult, error) {

	invoices, err := p.invoices()
	if err != nil {
		return ConfirmationResult{}, err
	}

	newM, err := invoices.membership(p.Order.UserIDs, p.Member)
	if err != nil {
		return ConfirmationResult{}, err
	}

	var archiver = reader.NewArchiver().WithOrderKind(p.Order.Kind)
	if p.Order.PaymentMethod == enum.PayMethodAli {
		archiver = archiver.ByAli()
	} else if p.Order.PaymentMethod == enum.PayMethodWx {
		archiver = archiver.ByWechat()
	}

	return ConfirmationResult{
		Payment: p.Payment,
		Order: p.Order.Confirmed(
			p.Payment.ConfirmedUTC,
			invoices.Purchased.TimeSlot),
		Invoices:   invoices,
		Membership: newM,
		Versioned: reader.NewMembershipVersioned(newM).
			WithPriorVersion(p.Member).
			ArchivedBy(archiver),
		Notify: true,
	}, nil
}
