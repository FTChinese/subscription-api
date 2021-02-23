package subs

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
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

// ConfirmationParams contains data used to confirm an order.
type ConfirmationParams struct {
	Payment PaymentResult
	Order   Order
	Member  reader.Membership
}

func (params ConfirmationParams) confirmOrder() (ConfirmedOrder, error) {
	switch params.Order.Kind {
	case enum.OrderKindCreate, enum.OrderKindRenew:
		return ConfirmedOrder{
			Order: params.Order.newOrRenewalConfirm(params.Payment.ConfirmedUTC, params.Member.ExpireDate),
			AddOn: addon.AddOn{},
		}, nil

	case enum.OrderKindUpgrade:
		if params.Member.Tier == enum.TierPremium {
			params.Order.Kind = enum.OrderKindRenew
			return ConfirmedOrder{
				Order: params.Order.newOrRenewalConfirm(params.Payment.ConfirmedUTC, params.Member.ExpireDate),
				AddOn: addon.AddOn{},
			}, nil
		}

		return ConfirmedOrder{
			Order: params.Order.upgradeConfirm(params.Payment.ConfirmedUTC),
			AddOn: params.Member.CarryOver(addon.CarryOverFromUpgrade).WithOrderID(params.Order.ID),
		}, nil

	case enum.OrderKindAddOn:
		params.Order.ConfirmedAt = params.Payment.ConfirmedUTC
		return ConfirmedOrder{
			Order: params.Order,
			AddOn: params.Order.ToAddOn(),
		}, nil

	default:
		return ConfirmedOrder{}, errors.New("unknown order kind")
	}
}

func (params ConfirmationParams) snapshot() reader.MemberSnapshot {
	if params.Member.IsZero() {
		return reader.MemberSnapshot{}
	}

	return params.Member.Snapshot(
		reader.FtcArchiver(params.Order.Kind))
}

// ConfirmedOrder contains the result of an order in confirmed state,
// together with optional add-on.
type ConfirmedOrder struct {
	Order Order       `json:"order"`
	AddOn addon.AddOn `json:"-"`
}

func newMembership(co ConfirmedOrder, currentMember reader.Membership) reader.Membership {
	// If an order is created as an add-on, only add the reserved days
	// to current membership.
	if co.Order.Kind == enum.OrderKindAddOn {
		return currentMember.
			WithReservedDays(co.AddOn.ToReservedDays())
	}

	return reader.Membership{
		MemberID:      co.Order.MemberID,
		Edition:       co.Order.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    co.Order.EndDate,
		PaymentMethod: co.Order.PaymentMethod,
		FtcPlanID:     null.StringFrom(co.Order.PlanID),
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        enum.SubsStatusNull,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		ReservedDays: currentMember.ReservedDays.
			Plus(co.AddOn.ToReservedDays()),
	}.Sync()
}

// ConfirmationResult contains all the data in the process of confirming an order.
// This is also used as the http response for manual confirmation.
type ConfirmationResult struct {
	Payment PaymentResult `json:"payment"` // Empty if order is already confirmed.
	ConfirmedOrder
	Membership reader.Membership     `json:"membership"` // The updated membership. Empty if order is already confirmed.
	Snapshot   reader.MemberSnapshot `json:"-"`
	Notify     bool                  `json:"-"`
}

func NewConfirmationResult(p ConfirmationParams) (ConfirmationResult, error) {
	co, err := p.confirmOrder()
	if err != nil {
		return ConfirmationResult{}, err
	}

	return ConfirmationResult{
		Payment:        p.Payment,
		ConfirmedOrder: co,
		Membership:     newMembership(co, p.Member),
		Snapshot:       p.snapshot(),
		Notify:         true,
	}, nil
}
