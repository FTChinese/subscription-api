package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/collection"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/product"
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
	product.Edition
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

func groupAddOns(l []AddOn) map[product.Edition][]AddOn {
	g := make(map[product.Edition][]AddOn)

	for _, v := range l {
		g[v.Edition] = append(g[v.Edition], v)
	}

	return g
}

func sumAddOns(addOns []AddOn) product.Duration {
	var cycles, days int64
	for _, v := range addOns {
		cycles += v.CycleCount
		days += v.DaysRemained
	}

	return product.Duration{
		CycleCount: cycles,
		ExtraDays:  days,
	}
}

func newMembershipFromAddOn(addOns []AddOn, m reader.Membership) reader.Membership {
	startTime := dt.PickLater(time.Now(), m.ExpireDate.Time)

	latestAddOn := addOns[0]

	dur := sumAddOns(addOns)
	dateRange := dt.NewDateRange(startTime).
		WithCycleN(latestAddOn.Cycle, int(dur.CycleCount)).
		AddDays(int(dur.ExtraDays))

	return reader.Membership{
		MemberID:      m.MemberID,
		Edition:       latestAddOn.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    dateRange.EndDate,
		PaymentMethod: latestAddOn.PaymentMethod,
		FtcPlanID:     latestAddOn.PlanID,
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        0,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		ReservedDays: reader.ReservedDays{
			Standard: m.ReservedDays.Standard,
			Premium:  0, // Clear premium days
		},
	}.Sync()
}

func TransferAddOn(addOns []AddOn, m reader.Membership) AddOnConsumed {
	g := groupAddOns(addOns)

	snapshot := m.Snapshot(reader.FtcArchiver(enum.OrderKindAddOn))
	var ids = collection.NewStringSet()

	prmAddOns, ok := g[product.PremiumEdition]
	if ok && len(prmAddOns) > 0 {
		collectAddOnIDs(ids, prmAddOns)
		return AddOnConsumed{
			AddOnIDs:   ids,
			Membership: newMembershipFromAddOn(prmAddOns, m),
			Snapshot:   snapshot,
		}
	}

	stdYearAddOns, ok := g[product.StdYearEdition]
	if ok && len(stdYearAddOns) > 0 {
		collectAddOnIDs(ids, stdYearAddOns)
		m = newMembershipFromAddOn(stdYearAddOns, m)
	}

	stdMonthAddOns, ok := g[product.StdMonthEdition]
	if ok && len(stdMonthAddOns) > 0 {
		collectAddOnIDs(ids, stdMonthAddOns)
		m = newMembershipFromAddOn(stdMonthAddOns, m)
	}

	return AddOnConsumed{
		AddOnIDs:   ids,
		Membership: m,
		Snapshot:   snapshot,
	}
}

func collectAddOnIDs(dest collection.StringSet, addOns []AddOn) {
	for _, v := range addOns {
		dest[v.ID] = nil
	}
}

type AddOnConsumed struct {
	AddOnIDs   collection.StringSet
	Membership reader.Membership
	Snapshot   reader.MemberSnapshot
}
