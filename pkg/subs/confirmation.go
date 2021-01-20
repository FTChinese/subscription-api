package subs

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
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

type ConfirmationParams struct {
	Payment PaymentResult
	Order   Order
	Member  reader.Membership
}

type PaymentConfirmed struct {
	Order    Order                 `json:"order"` // The confirmed order.
	AddOn    AddOn                 `json:"addOn"`
	Snapshot reader.MemberSnapshot `json:"-"` // Snapshot of previous membership
}

func NewPaymentConfirmed(p ConfirmationParams) (PaymentConfirmed, error) {
	snapshot := p.Member.Snapshot(reader.FtcArchiver(p.Order.Kind))
	if !p.Member.IsZero() {
		snapshot = snapshot.WithOrder(p.Order.ID)
	}

	switch p.Order.Kind {
	case enum.OrderKindCreate, enum.OrderKindRenew:
		return PaymentConfirmed{
			Order:    p.Order.newOrRenewalConfirm(p.Payment.ConfirmedUTC, p.Member.ExpireDate),
			AddOn:    AddOn{},
			Snapshot: snapshot,
		}, nil

	case enum.OrderKindUpgrade:
		if p.Member.Tier == enum.TierPremium {
			p.Order.Kind = enum.OrderKindRenew
			return PaymentConfirmed{
				Order:    p.Order.newOrRenewalConfirm(p.Payment.ConfirmedUTC, p.Member.ExpireDate),
				AddOn:    AddOn{},
				Snapshot: snapshot,
			}, nil
		}

		return PaymentConfirmed{
			Order:    p.Order.upgradeConfirm(p.Payment.ConfirmedUTC),
			AddOn:    NewUpgradeCarryOver(p.Order, p.Member),
			Snapshot: snapshot,
		}, nil

	case enum.OrderKindAddOn:
		p.Order.ConfirmedAt = p.Payment.ConfirmedUTC
		return PaymentConfirmed{
			Order:    p.Order,
			AddOn:    NewAddOn(p.Order),
			Snapshot: snapshot,
		}, nil

	default:
		return PaymentConfirmed{}, errors.New("unknown order kind")
	}
}

func MustConfirmPayment(p ConfirmationParams) PaymentConfirmed {
	c, err := NewPaymentConfirmed(p)

	if err != nil {
		panic(err)
	}

	return c
}

func NewMembership(p PaymentConfirmed) reader.Membership {
	// If an order is created as an add-on, only add the reserved days
	// to current membership.
	if p.Order.Kind == enum.OrderKindAddOn {
		m := p.Snapshot.Membership
		return m.WithReservedDays(p.AddOn.ToReservedDays())
	}

	return reader.Membership{
		MemberID:      p.Order.MemberID,
		Edition:       p.Order.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    p.Order.EndDate,
		PaymentMethod: p.Order.PaymentMethod,
		FtcPlanID:     null.StringFrom(p.Order.PlanID),
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        enum.SubsStatusNull,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		ReservedDays:  p.Snapshot.Membership.ReservedDays.Plus(p.AddOn.ToReservedDays()),
	}.Sync()
}

// ConfirmationResult contains all the data in the process of confirming an order.
// This is also used as the http response for manual confirmation.
type ConfirmationResult struct {
	Payment PaymentResult `json:"payment"` // Empty if order is already confirmed.
	PaymentConfirmed
	Membership reader.Membership `json:"membership"` // The updated membership. Empty if order is already confirmed.
	Notify     bool              `json:"-"`
}

func NewConfirmationResult(p ConfirmationParams) (ConfirmationResult, error) {
	pc, err := NewPaymentConfirmed(p)
	if err != nil {
		return ConfirmationResult{}, err
	}

	m := NewMembership(pc)

	return ConfirmationResult{
		Payment:          p.Payment,
		PaymentConfirmed: pc,
		Membership:       m,
		Notify:           true,
	}, nil
}
