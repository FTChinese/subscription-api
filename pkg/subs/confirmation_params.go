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
	Order   Order // The order not confirmed yet.
	Member  reader.Membership
}

// purchasedTimeParams collects the essential parameters used to
// determine the start and end date of an order.
func (p ConfirmationParams) purchasedTimeParams() PurchasedTimeParams {
	return PurchasedTimeParams{
		ConfirmedAt:    p.Payment.ConfirmedUTC,
		ExpirationDate: p.Member.ExpireDate,
		Date:           dt.NewYearMonthDay(p.Order.Cycle),
		OrderKind:      p.Order.CalibratedKind(p.Member),
	}
}

// purchaseInvoice creates an invoice from an order.
func (p ConfirmationParams) purchaseInvoice() (invoice.Invoice, error) {
	return NewOrderInvoice(p.purchasedTimeParams(), p.Order)
}

func (p ConfirmationParams) carryOverInvoice() invoice.Invoice {
	if p.Order.Kind == enum.OrderKindUpgrade {
		// Add order id to this carry over invoice so that later we could know which order caused carry-over.
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
