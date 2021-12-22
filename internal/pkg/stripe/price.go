package stripe

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"strconv"
	"time"
)

// PriceMetadata parsed the fields defined in stripe price's medata field.
// Those are customer-defined key-value pairs.
type PriceMetadata struct {
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
func NewPriceMeta(m map[string]string) PriceMetadata {
	tier, _ := enum.ParseTier(m["tier"])
	pd, _ := strconv.Atoi(m["period_days"])
	years, _ := strconv.Atoi(m["years"])
	months, _ := strconv.Atoi(m["months"])
	days, _ := strconv.Atoi(m["days"])
	isIntro, _ := strconv.ParseBool(m["introductory"])
	start, _ := m["start_utc"]
	end, _ := m["end_utc"]

	return PriceMetadata{
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

func (m PriceMetadata) SyncPeriod() PriceMetadata {
	if m.PeriodDays == 0 {
		m.PeriodDays = m.PeriodCount.TotalDays()
	}

	return m
}

func PriceMetaParams(p price.Price, isIntro bool) map[string]string {

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
		"introductory": strconv.FormatBool(isIntro),
		"start_utc":    start,
		"end_utc":      end,
	}
}

// PriceRecurring is the equivalence of stripe.PriceRecurring.
// Deprecated
type PriceRecurring struct {
	Interval      stripe.PriceRecurringInterval  `json:"interval"`
	IntervalCount int64                          `json:"intervalCount"`
	UsageType     stripe.PriceRecurringUsageType `json:"usageType"`
}

// NewPriceRecurring converts stripe recurring.
// Deprecated.
func NewPriceRecurring(r *stripe.PriceRecurring) PriceRecurring {
	if r == nil {
		return PriceRecurring{}
	}

	return PriceRecurring{
		Interval:      r.Interval,
		IntervalCount: r.IntervalCount,
		UsageType:     r.UsageType,
	}
}

func (r PriceRecurring) IsZero() bool {
	return r.Interval == "" && r.IntervalCount == 0 && r.UsageType == ""
}

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

type Price struct {
	ID             string           `json:"id"`
	Active         bool             `json:"active"`
	Currency       stripe.Currency  `json:"currency"`
	IsIntroductory bool             `json:"isIntroductory"`
	Kind           price.Kind       `json:"kind"`
	LiveMode       bool             `json:"liveMode"`
	Metadata       PriceMetadata    `json:"metadata"` // Deprecated
	Nickname       string           `json:"nickname"`
	Product        string           `json:"product"` // Deprecated
	ProductID      string           `json:"productId"`
	PeriodCount    dt.YearMonthDay  `json:"periodCount"`
	Recurring      PriceRecurring   `json:"recurring"` // Deprecated
	Tier           enum.Tier        `json:"tier"`      // The tier of this price.
	Type           stripe.PriceType `json:"type"`      // Deprecated
	UnitAmount     int64            `json:"unitAmount"`
	StartUTC       null.String      `json:"startUtc"` // Start time if Introductory is true; otherwise omit.
	EndUTC         null.String      `json:"endUtc"`   // End time if Introductory is true; otherwise omit.
	Created        int64            `json:"created"`
}

func NewPrice(p *stripe.Price) Price {

	meta := NewPriceMeta(p.Metadata)

	var period dt.YearMonthDay
	if p.Recurring == nil {
		period = meta.PeriodCount
	} else {
		period = parseRecurringPeriod(p.Recurring)
	}

	return Price{
		ID:             p.ID,
		Active:         p.Active,
		Currency:       p.Currency,
		IsIntroductory: meta.Introductory,
		Kind:           price.Kind(p.Type),
		LiveMode:       p.Livemode,
		Metadata:       meta,
		Nickname:       p.Nickname,
		Product:        p.Product.ID,
		ProductID:      p.Product.ID,
		PeriodCount:    period,
		Recurring:      NewPriceRecurring(p.Recurring),
		Tier:           meta.Tier,
		Type:           p.Type,
		UnitAmount:     p.UnitAmount,
		StartUTC:       meta.StartUTC,
		EndUTC:         meta.EndUTC,
		Created:        p.Created,
	}
}

func (p Price) IsZero() bool {
	return p.ID == ""
}

// Edition deduces the edition of stripe price since that
// information might be missing.
func (p Price) Edition() price.Edition {
	return price.Edition{
		Tier:  p.Tier,
		Cycle: p.PeriodCount.EqCycle(),
	}
}
