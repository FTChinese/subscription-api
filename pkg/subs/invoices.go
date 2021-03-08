package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

func NewOrderInvoice(timeParams PurchasedTimeParams, o Order) (invoice.Invoice, error) {
	timeRange, err := timeParams.Build()
	if err != nil {
		return invoice.Invoice{}, err
	}

	var addOnSource addon.Source
	if o.Kind == enum.OrderKindAddOn {
		addOnSource = addon.SourceUserPurchase
	}

	return invoice.Invoice{
		ID:             db.InvoiceID(),
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
		ConsumedUTC:    timeParams.ConfirmedAt,
		DateTimePeriod: timeRange.ToDateTimePeriod(),
		CarriedOverUtc: chrono.Time{},
	}, nil
}

type Invoices struct {
	Purchased   invoice.Invoice
	CarriedOver invoice.Invoice // If carry-over invoice exists, you should add timestamp to old invoices that are not yet completely consumed.
}

// Create new membership based on purchased invoice and optional carry over invoice.
func (i Invoices) membership(userID reader.MemberID, current reader.Membership) (reader.Membership, error) {
	// This will update expiration date.
	current, err := i.Purchased.NewMembership(userID, current)

	if err != nil {
		return reader.Membership{}, err
	}

	if i.CarriedOver.IsZero() {
		return current, nil
	}

	// This will update add-on part if carried over invoice is not empty.
	return i.CarriedOver.NewMembership(userID, current)
}
