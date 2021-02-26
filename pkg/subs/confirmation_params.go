package subs

import (
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// ConfirmationParams contains data used to confirm an order.
type ConfirmationParams struct {
	Payment PaymentResult
	Order   Order
	Member  reader.Membership
}

func (params ConfirmationParams) confirmNewOrRenewalOrder() Order {

	// Prevent duplicate confirmation
	if params.Order.IsConfirmed() {
		return params.Order
	}

	params.Order.ConfirmedAt = params.Payment.ConfirmedUTC

	startTime := dt.PickLater(params.Payment.ConfirmedUTC.Time, params.Member.ExpireDate.Time)

	params.Order.DatePeriod = dt.NewTimeRange(startTime).
		WithCycle(params.Order.Cycle).
		AddDays(trialDays).
		ToDatePeriod()

	return params.Order
}

func (params ConfirmationParams) confirmUpgradeOrder() Order {
	// Prevent duplicate confirmation.
	if params.Order.IsConfirmed() {
		return params.Order
	}

	params.Order.ConfirmedAt = params.Payment.ConfirmedUTC
	params.Order.DatePeriod = dt.NewTimeRange(params.Payment.ConfirmedUTC.Time).
		WithCycle(params.Order.Cycle).
		AddDays(trialDays).
		ToDatePeriod()

	return params.Order
}

func (params ConfirmationParams) snapshot() reader.MemberSnapshot {
	if params.Member.IsZero() {
		return reader.MemberSnapshot{}
	}

	return params.Member.Snapshot(
		reader.FtcArchiver(params.Order.Kind))
}
