package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// ConfirmationParams contains data used to invoice an order.
type ConfirmationParams struct {
	Payment PaymentResult
	Order   Order
	Member  reader.Membership
}

func (p ConfirmationParams) purchasedTimeParams() PurchasedTimeParams {
	return PurchasedTimeParams{
		ConfirmedAt:    p.Payment.ConfirmedUTC,
		ExpirationDate: p.Member.ExpireDate,
		Date:           dt.NewYearMonthDay(p.Order.Cycle),
		OrderKind:      p.Order.CalibratedKind(p.Member),
	}
}

func (p ConfirmationParams) purchaseInvoice() (invoice.Invoice, error) {
	return NewOrderInvoice(p.purchasedTimeParams(), p.Order)
}

func (p ConfirmationParams) carryOverInvoice() invoice.Invoice {
	if p.Order.Kind == enum.OrderKindUpgrade {
		// We have to add this invoice's days to current membership's addon part.
		return p.Member.CarryOverInvoice().
			WithOrderID(p.Order.ID)
	}

	return invoice.Invoice{}
}

// Build Invoice for when confirming an order,
// and optionally create a carry-over invoice for
// upgrading.
func (p ConfirmationParams) invoices() (Invoices, error) {
	purchased, err := p.purchaseInvoice()
	if err != nil {
		return Invoices{}, err
	}

	return Invoices{
		Purchased:   purchased,
		CarriedOver: p.carryOverInvoice(),
	}, nil
}

func (p ConfirmationParams) snapshot() reader.MemberSnapshot {
	return p.Member.Snapshot(reader.FtcArchiver(p.Order.Kind))
}
