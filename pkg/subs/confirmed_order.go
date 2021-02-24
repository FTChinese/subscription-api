package subs

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

// ConfirmedOrder contains the result of an order in confirmed state,
// together with optional add-on.
type ConfirmedOrder struct {
	Order Order       `json:"order"`
	AddOn addon.AddOn `json:"-"`
}

// NewConfirmedOrder confirms an order based on payment result and
// current membership.
func NewConfirmedOrder(params ConfirmationParams) (ConfirmedOrder, error) {
	switch params.Order.Kind {
	case enum.OrderKindCreate, enum.OrderKindRenew:
		return ConfirmedOrder{
			Order: params.confirmNewOrRenewalOrder(),
			AddOn: addon.AddOn{},
		}, nil

	case enum.OrderKindUpgrade:
		if params.Member.Tier == enum.TierPremium {
			params.Order.Kind = enum.OrderKindRenew
			return ConfirmedOrder{
				Order: params.confirmNewOrRenewalOrder(),
				AddOn: addon.AddOn{},
			}, nil
		}

		return ConfirmedOrder{
			Order: params.confirmUpgradeOrder(),
			AddOn: params.Member.
				CarryOver(addon.CarryOverFromUpgrade).
				WithOrderID(params.Order.ID),
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

// newMembership creates a updates current membership based on
// a one-time purchase.
func (co ConfirmedOrder) newMembership(currentMember reader.Membership) reader.Membership {
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
