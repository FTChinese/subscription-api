package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/stripe/stripe-go/v72"
	"strconv"
	"time"
)

// StripePriceMeta parsed the fields defined in stripe price's medata field.
// Those are customer-defined key-value pairs.
type StripePriceMeta struct {
	Introductory bool            `json:"introductory"` // Is this an introductory price? If false, it is treated as regular price.
	PeriodDays   int64           `json:"periodDays"`   // Deprecated
	PeriodCount  dt.YearMonthDay `json:"-"`
	Tier         enum.Tier       `json:"tier"`     // The tier of this price.
	StartUTC     chrono.Time     `json:"startUtc"` // Start time if Introductory is true; otherwise omit.
	EndUTC       chrono.Time     `json:"endUtc"`   // End time if Introductory is true; otherwise omit.
}

func BuildStripePriceMeta(p FtcPrice) map[string]string {
	start := ""
	end := ""
	if !p.StartUTC.IsZero() {
		start = p.StartUTC.In(time.UTC).Format(time.RFC3339)
	}

	if !p.EndUTC.IsZero() {
		end = p.EndUTC.In(time.UTC).Format(time.RFC3339)
	}

	return map[string]string{
		"tier":         p.Tier.String(),
		"years":        strconv.FormatInt(p.PeriodCount.Years, 10),
		"months":       strconv.FormatInt(p.PeriodCount.Months, 10),
		"days":         strconv.FormatInt(p.PeriodCount.Days, 10),
		"introductory": strconv.FormatBool(p.IsOneTime()),
		"start_utc":    start,
		"end_utc":      end,
	}
}

// ParseStripePriceMeta converts stripe price metadata to a struct.
// - tier: standard | premium
// - years: number
// - months: number
// - days: number
// - introductory: boolean
// - start_utc?: string
// - end_utc?: string
func ParseStripePriceMeta(m map[string]string) StripePriceMeta {
	tier, _ := enum.ParseTier(m["tier"])
	pd, _ := strconv.Atoi(m["period_days"])
	years, _ := strconv.Atoi(m["years"])
	months, _ := strconv.Atoi(m["months"])
	days, _ := strconv.Atoi(m["days"])
	isIntro, _ := strconv.ParseBool(m["introductory"])
	start, _ := m["start_utc"]
	end, _ := m["end_utc"]

	startTime, _ := time.Parse(time.RFC3339, start)
	endTime, _ := time.Parse(time.RFC3339, end)

	return StripePriceMeta{
		Tier:       tier,
		PeriodDays: int64(pd),
		PeriodCount: dt.YearMonthDay{
			Years:  int64(years),
			Months: int64(months),
			Days:   int64(days),
		},
		Introductory: isIntro,
		StartUTC:     chrono.TimeFrom(startTime),
		EndUTC:       chrono.TimeFrom(endTime),
	}.SyncPeriod()
}

func (m StripePriceMeta) SyncPeriod() StripePriceMeta {
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
	IsFromStripe   bool               `json:"-"`
	ID             string             `json:"id" db:"id"`
	Active         bool               `json:"active" db:"active"`
	Currency       stripe.Currency    `json:"currency" db:"currency"`
	IsIntroductory bool               `json:"isIntroductory" db:"is_intro"` // Deprecated
	Kind           Kind               `json:"kind" db:"kind"`
	LiveMode       bool               `json:"liveMode" db:"live_mode"`
	Nickname       string             `json:"nickname" db:"nickname"`
	ProductID      string             `json:"productId" db:"product_id"`
	PeriodCount    ColumnYearMonthDay `json:"periodCount" db:"period_count"`
	Tier           enum.Tier          `json:"tier" db:"tier"` // The tier of this price.
	UnitAmount     int64              `json:"unitAmount" db:"unit_amount"`
	StartUTC       chrono.Time        `json:"startUtc" db:"start_utc"` // Start time if Introductory is true; otherwise omit.
	EndUTC         chrono.Time        `json:"endUtc" db:"end_utc"`     // End time if Introductory is true; otherwise omit.
	Created        int64              `json:"created" db:"created"`

	Metadata  StripePriceMeta      `json:"metadata"`  // Deprecated
	Product   string               `json:"product"`   // Deprecated. Use ProductID
	Recurring StripePriceRecurring `json:"recurring"` // Deprecated
	Type      stripe.PriceType     `json:"type"`      // Deprecated. Use Kind.
}

func NewStripePrice(p *stripe.Price) StripePrice {

	meta := ParseStripePriceMeta(p.Metadata)

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
		PeriodCount:    ColumnYearMonthDay{period},
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

func (p StripePrice) IsIntro() bool {
	return p.Kind == KindOneTime
}

// Edition deduces the edition of stripe price since that
// information might be missing.
func (p StripePrice) Edition() Edition {
	return Edition{
		Tier:  p.Tier,
		Cycle: p.PeriodCount.EqCycle(),
	}
}
