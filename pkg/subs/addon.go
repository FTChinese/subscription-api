package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
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
	CycleCount    int64          `json:"cycleCount" db:"cycle_count"`
	DaysRemained  int64          `json:"daysRemained" db:"days_remained"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	OrderID       null.String    `json:"orderId" db:"order_id"`
	CompoundID    string         `json:"compoundId" db:"compound_id"`
	CreatedUTC    chrono.Time    `json:"createdUtc" db:"created_utc"`
	ConsumedUTC   chrono.Time    `json:"consumedUtc" db:"consumed_utc"`
}

func NewAddOn(o Order) AddOn {
	return AddOn{
		ID:            db.AddOnID(),
		Edition:       o.Edition,
		CycleCount:    1,
		DaysRemained:  trialDays,
		PaymentMethod: o.PaymentMethod,
		OrderID:       null.StringFrom(o.ID),
		CompoundID:    o.CompoundID,
		CreatedUTC:    chrono.TimeNow(),
		ConsumedUTC:   chrono.Time{},
	}
}

// NewUpgradeAddOn moves the remaining days of a standard subscription
// to addon portion upon upgrading to premium.
func NewUpgradeAddOn(o Order, m reader.Membership) AddOn {
	return AddOn{
		ID:            db.AddOnID(),
		Edition:       m.Edition,
		CycleCount:    0,
		DaysRemained:  m.RemainingDays(),
		PaymentMethod: m.PaymentMethod,
		OrderID:       null.StringFrom(o.ID), // Which order caused the current membership to move remaining days to reserved state.
		CompoundID:    o.CompoundID,
		CreatedUTC:    chrono.TimeNow(),
		ConsumedUTC:   chrono.Time{},
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
		FtcPlanID:     null.String{}, // TODO: carry over plan id from order
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
	}
}

func TransferAddOn(addOns []AddOn, m reader.Membership) AddOnConsumed {
	g := groupAddOns(addOns)

	snapshot := m.Snapshot(reader.FtcArchiver(enum.OrderKindAddOn))

	prmAddOns, ok := g[product.PremiumEdition]
	if ok && len(prmAddOns) > 0 {
		return AddOnConsumed{
			AddOns:     prmAddOns,
			Membership: newMembershipFromAddOn(prmAddOns, m),
			Snapshot:   snapshot,
		}
	}

	var consumed []AddOn

	stdYearAddOns, ok := g[product.StdYearEdition]
	if ok && len(stdYearAddOns) > 0 {
		consumed = append(addOns, stdYearAddOns...)
		m = newMembershipFromAddOn(stdYearAddOns, m)
	}

	stdMonthAddOns, ok := g[product.StdMonthEdition]
	if ok && len(stdMonthAddOns) > 0 {
		consumed = append(addOns, stdMonthAddOns...)
		m = newMembershipFromAddOn(stdMonthAddOns, m)
	}

	return AddOnConsumed{
		AddOns:     consumed,
		Membership: m,
		Snapshot:   snapshot,
	}
}

type AddOnConsumed struct {
	AddOns     []AddOn
	Membership reader.Membership
	Snapshot   reader.MemberSnapshot
}

func GetAddOnIDs(addOns []AddOn) []string {
	ids := make([]string, 0)

	for _, v := range addOns {
		ids = append(ids, v.ID)
	}

	return ids
}
