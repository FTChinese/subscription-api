package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

// ConfirmationParams contains data used to invoice an order.
type ConfirmationParams struct {
	Payment PaymentResult
	Order   Order
	Member  reader.Membership
}

func (p ConfirmationParams) invoice() (invoice.Invoice, error) {
	return p.Order.invoice(p.Payment.ConfirmedUTC, p.Member.ExpireDate)
}

func (p ConfirmationParams) membership(inv invoice.Invoice) reader.Membership {
	return newMembership(p.Order.MemberID, p.Member, inv)
}

func (p ConfirmationParams) confirmedOrder(period dt.DateTimePeriod) Order {
	p.Order.ConfirmedAt = p.Payment.ConfirmedUTC
	p.Order.DatePeriod = period.ToDatePeriod()

	return p.Order
}

func (p ConfirmationParams) snapshot() reader.MemberSnapshot {
	return p.Member.Snapshot(reader.FtcArchiver(p.Order.Kind))
}

func newMembership(userID reader.MemberID, m reader.Membership, inv invoice.Invoice) reader.Membership {
	if inv.OrderKind == enum.OrderKindAddOn {
		return m.WithReservedDays(inv.ToReservedDays())
	}

	// If the invoice is not intended for add-on, it must have period set.
	return reader.Membership{
		MemberID:      userID, // TODO: user id from order.
		Edition:       inv.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    chrono.DateFrom(inv.EndUTC.Time),
		PaymentMethod: inv.PaymentMethod,
		FtcPlanID:     inv.PriceID,
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        enum.SubsStatusNull,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		ReservedDays:  m.ReservedDays,
	}.Sync()
}
