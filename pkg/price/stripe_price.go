package price

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"strconv"
)

// StripePriceMetadata parsed the fields defined in stripe price's medata field.
// Those are customer-defined key-value pairs.
type StripePriceMetadata struct {
	Introductory bool            `json:"introductory"` // Is this an introductory price? If false, it is treated as regular price.
	PeriodDays   int64           `json:"periodDays"`
	PeriodCount  dt.YearMonthDay `json:"-"`
	Tier         enum.Tier       `json:"tier"`     // The tier of this price.
	StartUTC     null.String     `json:"startUtc"` // Start time if Introductory is true; otherwise omit.
	EndUTC       null.String     `json:"endUtc"`   // End time if Introductory is true; otherwise omit.
}

// NewPriceMeta converts stripe price metadata to a struct.
// - tier: standard | premium
// - years: number
// - months: number
// - days: number
// - introductory: boolean
// - start_utc?: string
// - end_utc?: string
func NewPriceMeta(m map[string]string) StripePriceMetadata {
	tier, _ := enum.ParseTier(m["tier"])
	pd, _ := strconv.Atoi(m["period_days"])
	years, _ := strconv.Atoi(m["years"])
	months, _ := strconv.Atoi(m["months"])
	days, _ := strconv.Atoi(m["days"])
	isIntro, _ := strconv.ParseBool(m["introductory"])
	start, _ := m["start_utc"]
	end, _ := m["end_utc"]

	return StripePriceMetadata{
		Tier:       tier,
		PeriodDays: int64(pd),
		PeriodCount: dt.YearMonthDay{
			Years:  int64(years),
			Months: int64(months),
			Days:   int64(days),
		},
		Introductory: isIntro,
		StartUTC:     null.NewString(start, start != ""),
		EndUTC:       null.NewString(end, end != ""),
	}.SyncPeriod()
}

func (m StripePriceMetadata) SyncPeriod() StripePriceMetadata {
	if m.PeriodDays == 0 {
		m.PeriodDays = m.PeriodCount.TotalDays()
	}

	return m
}

// StripePriceRecurring is the equivalence of stripe.PriceRecurring.
// Deprecated
type StripePriceRecurring struct {
	Interval      stripe.PriceRecurringInterval  `json:"interval"`
	IntervalCount int64                          `json:"intervalCount"`
	UsageType     stripe.PriceRecurringUsageType `json:"usageType"`
}

// NewPriceRecurring converts stripe recurring.
// Deprecated.
func NewPriceRecurring(r *stripe.PriceRecurring) StripePriceRecurring {
	if r == nil {
		return StripePriceRecurring{}
	}

	return StripePriceRecurring{
		Interval:      r.Interval,
		IntervalCount: r.IntervalCount,
		UsageType:     r.UsageType,
	}
}

func (r StripePriceRecurring) IsZero() bool {
	return r.Interval == "" && r.IntervalCount == 0 && r.UsageType == ""
}

// parseRecurringPeriod tries to deduce a dt.YearMonthDay from stripe.PriceRecurring
// when it is not set in stripe's Metadata field
func parseRecurringPeriod(r *stripe.PriceRecurring) dt.YearMonthDay {
	if r == nil {
		return dt.YearMonthDay{}
	}

	ymd := dt.YearMonthDay{}

	switch r.Interval {
	case stripe.PriceRecurringIntervalYear:
		ymd.Years = r.IntervalCount

	case stripe.PriceRecurringIntervalMonth:
		ymd.Months = r.IntervalCount

	case stripe.PriceRecurringIntervalWeek:
		ymd.Days = r.IntervalCount * 7

	case stripe.PriceRecurringIntervalDay:
		ymd.Days = r.IntervalCount
	}

	return ymd
}

type StripePrice struct {
	IsFromStripe   bool            `json:"-"`
	ID             string          `json:"id"`
	Active         bool            `json:"active"`
	Currency       stripe.Currency `json:"currency"`
	IsIntroductory bool            `json:"isIntroductory"`
	Kind           Kind            `json:"kind"`
	LiveMode       bool            `json:"liveMode"`
	Nickname       string          `json:"nickname"`
	ProductID      string          `json:"productId"`
	PeriodCount    dt.YearMonthDay `json:"periodCount"`
	Tier           enum.Tier       `json:"tier"` // The tier of this price.
	UnitAmount     int64           `json:"unitAmount"`
	StartUTC       null.String     `json:"startUtc"` // Start time if Introductory is true; otherwise omit.
	EndUTC         null.String     `json:"endUtc"`   // End time if Introductory is true; otherwise omit.
	Created        int64           `json:"created"`

	Metadata  StripePriceMetadata  `json:"metadata"`  // Deprecated
	Product   string               `json:"product"`   // Deprecated. Use ProductID
	Recurring StripePriceRecurring `json:"recurring"` // Deprecated
	Type      stripe.PriceType     `json:"type"`      // Deprecated. Use Kind.
}

func NewPrice(p *stripe.Price) StripePrice {

	meta := NewPriceMeta(p.Metadata)

	var period dt.YearMonthDay
	if p.Recurring == nil {
		period = meta.PeriodCount
	} else {
		period = parseRecurringPeriod(p.Recurring)
	}

	return StripePrice{
		IsFromStripe:   true,
		ID:             p.ID,
		Active:         p.Active,
		Currency:       p.Currency,
		IsIntroductory: meta.Introductory,
		Kind:           Kind(p.Type),
		LiveMode:       p.Livemode,
		Nickname:       p.Nickname,
		ProductID:      p.Product.ID,
		PeriodCount:    period,
		Tier:           meta.Tier,
		UnitAmount:     p.UnitAmount,
		StartUTC:       meta.StartUTC,
		EndUTC:         meta.EndUTC,
		Created:        p.Created,

		Metadata:  meta,
		Product:   p.Product.ID,
		Recurring: NewPriceRecurring(p.Recurring),
		Type:      p.Type,
	}
}

func (p StripePrice) IsZero() bool {
	return p.ID == ""
}

// Edition deduces the edition of stripe price since that
// information might be missing.
func (p StripePrice) Edition() Edition {
	return Edition{
		Tier:  p.Tier,
		Cycle: p.PeriodCount.EqCycle(),
	}
}
