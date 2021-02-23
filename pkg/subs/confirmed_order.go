package subs

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/addon"
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
