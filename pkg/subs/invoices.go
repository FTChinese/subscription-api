package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

// NewOrderInvoice creates a new invoice from a unconfirmed order.
func NewOrderInvoice(timeParams PurchasedTimeParams, o Order) (invoice.Invoice, error) {
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
		ID:             pkg.InvoiceID(),
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

type Invoices struct {
	Purchased   invoice.Invoice
	CarriedOver invoice.Invoice // If carry-over invoice exists, you should add timestamp to old invoices that are not yet completely consumed.
}

// Create new membership based on purchased invoice and optional carry over invoice.
func (i Invoices) membership(userID pkg.UserIDs, current reader.Membership) (reader.Membership, error) {
	// This will update expiration date.
	current, err := current.WithInvoice(userID, i.Purchased)

	if err != nil {
		return reader.Membership{}, err
	}

	if i.CarriedOver.IsZero() {
		return current, nil
	}

	// This will update add-on part if carried over invoice is not empty.
	return current.WithInvoice(userID, i.CarriedOver)
}
