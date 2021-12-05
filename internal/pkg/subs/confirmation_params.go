package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
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
	return newOrderInvoice(p.purchasedTimeParams(), p.Order)
}

// newOrderInvoice creates a new invoice from a unconfirmed order.
// For addon order, the invoice do not have starting and ending time.
func newOrderInvoice(timeParams PurchasedTimeParams, o Order) (invoice.Invoice, error) {
	timeRange, err := timeParams.Build()
	if err != nil {
		return invoice.Invoice{}, err
	}

	var addOnSource addon.Source
	confirmedAt := timeParams.ConfirmedAt
	if o.Kind == enum.OrderKindAddOn {
		addOnSource = addon.SourceUserPurchase
		confirmedAt = chrono.Time{}
	}

	return invoice.Invoice{
		ID:             ids.InvoiceID(),
		CompoundID:     o.CompoundID,
		Edition:        o.Edition,
		YearMonthDay:   timeParams.Date,
		AddOnSource:    addOnSource,
		OrderID:        null.StringFrom(o.ID),
		OrderKind:      timeParams.OrderKind, // Note: use the calibrated order kind.
		PaidAmount:     o.Amount,
		PaymentMethod:  o.PaymentMethod,
		PriceID:        null.StringFrom(o.PlanID),
		CreatedUTC:     chrono.TimeNow(),
		ConsumedUTC:    confirmedAt,
		DateTimePeriod: timeRange.ToDateTimePeriod(),
		CarriedOverUtc: chrono.Time{},
	}, nil
}

// carryOverInvoice turns the remaining days of current membership into an invoice.
func (p ConfirmationParams) carryOverInvoice() invoice.Invoice {
	if p.Order.Kind == enum.OrderKindUpgrade {
		// Add order id to this carry over invoice so that later we could know which order caused carry-over.
		return p.Member.CarryOverInvoice().
			WithOrderID(p.Order.ID)
	}

	return invoice.Invoice{}
}

// invoices Build invoice when confirming an order,
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
	return p.Member.Snapshot(reader.NewOrderArchiver(p.Order.Kind))
}
