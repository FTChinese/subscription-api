package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/collection"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"time"
)

var cycleDays = map[enum.Cycle]int64{
	enum.CycleYear:  366,
	enum.CycleMonth: 31,
}

type AddOn struct {
	ID string `json:"id" db:"id"`
	price.Edition
	CycleCount         int64          `json:"cycleCount" db:"cycle_count"`
	DaysRemained       int64          `json:"daysRemained" db:"days_remained"`
	IsUpgradeCarryOver bool           `json:"isUpgradeCarryOver" db:"is_upgrade_carry_over"`
	PaymentMethod      enum.PayMethod `json:"payMethod" db:"payment_method"`
	CompoundID         string         `json:"compoundId" db:"compound_id"`
	OrderID            null.String    `json:"orderId" db:"order_id"`
	PlanID             null.String    `json:"planId" db:"plan_id"`
	CreatedUTC         chrono.Time    `json:"createdUtc" db:"created_utc"`
	ConsumedUTC        chrono.Time    `json:"consumedUtc" db:"consumed_utc"`
}

func NewAddOn(o Order) AddOn {
	return AddOn{
		ID:                 db.AddOnID(),
		Edition:            o.Edition,
		CycleCount:         1,
		DaysRemained:       trialDays,
		IsUpgradeCarryOver: false,
		PaymentMethod:      o.PaymentMethod,
		CompoundID:         o.CompoundID,
		OrderID:            null.StringFrom(o.ID),
		PlanID:             null.StringFrom(o.PlanID),
		CreatedUTC:         chrono.TimeNow(),
		ConsumedUTC:        chrono.Time{},
	}
}

// NewUpgradeCarryOver moves the remaining days of a standard subscription
// to addon portion upon upgrading to premium, and it will be
// restored when
func NewUpgradeCarryOver(o Order, m reader.Membership) AddOn {
	return AddOn{
		ID:                 db.AddOnID(),
		Edition:            m.Edition,
		CycleCount:         0,
		DaysRemained:       m.RemainingDays(),
		IsUpgradeCarryOver: true,
		PaymentMethod:      m.PaymentMethod,
		CompoundID:         o.CompoundID,
		OrderID:            null.StringFrom(o.ID), // Which order caused the current membership to move remaining days to reserved state.
		PlanID:             m.FtcPlanID,           // Save it so that it could be restored together with the remaining days.
		CreatedUTC:         chrono.TimeNow(),
		ConsumedUTC:        chrono.Time{},
	}
}

func (a AddOn) IsZero() bool {
	return a.ID == ""
}

func (a AddOn) GetDays() int64 {
	return a.CycleCount*cycleDays[a.Cycle] + a.DaysRemained
}

func (a AddOn) ToReservedDays() reader.ReservedDays {
	switch a.Tier {
	case enum.TierStandard:
		return reader.ReservedDays{
			Standard: a.GetDays(),
			Premium:  0,
		}
	case enum.TierPremium:
		return reader.ReservedDays{
			Standard: 0,
			Premium:  a.GetDays(),
		}

	default:
		return reader.ReservedDays{}
	}
}

type AddOnSum struct {
	Years  int
	Months int
	Days   int
	Latest AddOn
}

func (s AddOnSum) Membership(m reader.Membership) reader.Membership {
	startTime := dt.PickLater(time.Now(), m.ExpireDate.Time)

	endTime := startTime.AddDate(s.Years, s.Months, s.Days)

	// Clear the migrated days.
	reserved := m.ReservedDays
	switch s.Latest.Tier {
	case enum.TierStandard:
		reserved.Standard = 0

	case enum.TierPremium:
		reserved.Premium = 0
	}

	return reader.Membership{
		MemberID:      m.MemberID,
		Edition:       s.Latest.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    chrono.DateFrom(endTime),
		PaymentMethod: s.Latest.PaymentMethod,
		FtcPlanID:     s.Latest.PlanID,
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        0,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		ReservedDays:  reserved,
	}.Sync()
}

func GroupAddOns(addOns []AddOn) map[enum.Tier][]AddOn {
	g := make(map[enum.Tier][]AddOn)

	for _, v := range addOns {
		g[v.Tier] = append(g[v.Tier], v)
	}

	return g
}

func ReduceAddOns(addOns []AddOn) AddOnSum {
	if len(addOns) == 0 {
		return AddOnSum{}
	}

	sum := AddOnSum{
		Latest: addOns[0],
	}

	for _, v := range addOns {
		switch v.Cycle {
		case enum.CycleYear:
			sum.Years += int(v.CycleCount)

		case enum.CycleMonth:
			sum.Months += int(v.CycleCount)
		}

		sum.Days += int(v.DaysRemained)
	}

	return sum
}

func CollectAddOnIDs(addOns ...[]AddOn) collection.StringSet {
	dest := make(collection.StringSet)

	for _, a := range addOns {
		for _, v := range a {
			dest[v.ID] = nil
		}
	}

	return dest
}

type AddOnConsumed struct {
	AddOnIDs   collection.StringSet
	Membership reader.Membership
	Snapshot   reader.MemberSnapshot
}

func ConsumeAddOns(addOns []AddOn, m reader.Membership) AddOnConsumed {
	g := GroupAddOns(addOns)

	snapshot := m.Snapshot(reader.FtcArchiver(enum.OrderKindAddOn))

	prmAddOns, ok := g[enum.TierPremium]
	if ok && len(prmAddOns) > 0 {
		return AddOnConsumed{
			AddOnIDs:   CollectAddOnIDs(prmAddOns),
			Membership: ReduceAddOns(prmAddOns).Membership(m),
			Snapshot:   snapshot,
		}
	}

	stdAddOns, ok := g[enum.TierStandard]
	if ok && len(stdAddOns) > 0 {
		return AddOnConsumed{
			AddOnIDs:   CollectAddOnIDs(stdAddOns),
			Membership: ReduceAddOns(stdAddOns).Membership(m),
			Snapshot:   snapshot,
		}
	}

	return AddOnConsumed{
		Membership: m,
	}
}
