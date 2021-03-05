package reader

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/collection"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/guregu/null"
	"time"
)

type AddOnConsumed struct {
	AddOnIDs   collection.StringSet
	Membership Membership
	Snapshot   MemberSnapshot
}

func (m Membership) WithReservedDays(days addon.ReservedDays) Membership {
	m.ReservedDays = m.ReservedDays.Plus(days)
	return m
}

func (m Membership) WithAddOn(addOn addon.AddOn) Membership {
	m.ReservedDays = m.ReservedDays.Plus(addOn.ToReservedDays())
	return m
}

func (m Membership) HasAddOns() bool {
	return m.Standard > 0 || m.Premium > 0
}

func (m Membership) ShouldUseAddOn() error {
	if m.IsZero() {
		return errors.New("subscription backup days only applicable to an existing membership")
	}

	if !m.IsExpired() {
		return errors.New("backup days come into effect only after current subscription expired")
	}

	if !m.HasAddOns() {
		return errors.New("current membership does not have backup days")
	}

	return nil
}

func (m Membership) CarryOver(source addon.Source) addon.AddOn {
	return addon.AddOn{
		ID:              db.AddOnID(),
		Edition:         m.Edition,
		CycleCount:      0,
		DaysRemained:    m.RemainingDays(),
		CarryOverSource: source,
		PaymentMethod:   m.PaymentMethod,
		CompoundID:      m.CompoundID,
		OrderID:         null.String{},
		PlanID:          m.FtcPlanID, // Save it so that it could be restored together with the remaining days.
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}
}

// WithAddOnSum extends membership's expiration time by adding up all add-ons.
func (m Membership) WithAddOnSum(sum addon.Sum) Membership {
	startTime := dt.PickLater(time.Now(), m.ExpireDate.Time)

	endTime := startTime.AddDate(sum.Years, sum.Months, sum.Days)

	return Membership{
		MemberID:      m.MemberID,
		Edition:       sum.Latest.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    chrono.DateFrom(endTime),
		PaymentMethod: sum.Latest.PaymentMethod,
		FtcPlanID:     sum.Latest.PlanID,
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        0,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		ReservedDays:  m.ReservedDays.Clear(sum.Latest.Tier),
	}.Sync()
}

// ConsumeAddOns transfer addon periods to current membership.
func (m Membership) ConsumeAddOns(addOns []addon.AddOn) AddOnConsumed {
	addOnSums := addon.GroupAndReduce(addOns)

	snapshot := m.Snapshot(FtcArchiver(enum.OrderKindAddOn))

	prmSum, ok := addOnSums[enum.TierPremium]
	if ok {
		return AddOnConsumed{
			AddOnIDs:   prmSum.IDs,
			Membership: m.WithAddOnSum(prmSum),
			Snapshot:   snapshot,
		}
	}

	stdSum, ok := addOnSums[enum.TierStandard]
	if ok {
		return AddOnConsumed{
			AddOnIDs:   stdSum.IDs,
			Membership: m.WithAddOnSum(stdSum),
			Snapshot:   snapshot,
		}
	}

	return AddOnConsumed{
		Membership: m,
	}
}
