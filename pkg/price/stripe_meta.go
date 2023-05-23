package price

import (
	"strconv"
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/stripe/stripe-go/v72"
)

// StripePriceMeta parsed the fields defined in stripe price's medata field.
// Those are customer-defined key-value pairs.
// StartUTC and EndUTC only exists when
// Introductory is true.
type StripePriceMeta struct {
	Introductory bool            `json:"introductory"` // Is it an introductory price? Co-exist with StartUTC and EndUTC.
	PeriodCount  dt.YearMonthDay `json:"periodCount"`
	Tier         enum.Tier       `json:"tier"`     // The tier of this price. Always exists.
	StartUTC     chrono.Time     `json:"startUtc"` // Start time if Introductory is true; otherwise omit.
	EndUTC       chrono.Time     `json:"endUtc"`   // End time if Introductory is true; otherwise omit.

	PeriodDays int64 `json:"periodDays"` // Deprecated
}

func (m *StripePriceMeta) Validate() *render.ValidationError {
	// Tier is always required.
	if m.Tier == enum.TierNull {
		return &render.ValidationError{
			Message: "missing tier",
			Field:   "tier",
			Code:    render.CodeMissingField,
		}
	}

	// Recurring price.
	if !m.Introductory {
		m.PeriodCount = dt.YearMonthDay{}
		m.StartUTC = chrono.TimeZero()
		m.EndUTC = chrono.TimeZero()
		return nil
	}

	// Introductory price requires start and end time.
	if m.StartUTC.Time.IsZero() {
		return &render.ValidationError{
			Message: "missing start time",
			Field:   "startUtc",
			Code:    render.CodeMissingField,
		}
	}

	if m.EndUTC.Time.IsZero() {
		return &render.ValidationError{
			Message: "missing start time",
			Field:   "startUtc",
			Code:    render.CodeMissingField,
		}
	}

	return nil
}

// WithRecurring updates metadata based on stripe's recurring field.
// The first time a stripe price is save to ftc's db,
// chances are that no metadata fields set.
// In such case we could only deduce from a price's
// existing field.
func (m StripePriceMeta) WithRecurring(r *stripe.PriceRecurring) StripePriceMeta {
	if r != nil {
		// A subscripiton
		m.PeriodCount = parseRecurringPeriod(r)
		m.Introductory = false
		// Recurring prices do not have start end end time.
		m.StartUTC = chrono.TimeZero()
		m.EndUTC = chrono.TimeZero()
		return m
	}

	// Probably an introductory offer.
	// In such case, PeriodCount, StartUTC and EndUTC
	// could only be provided by us.
	m.Introductory = true
	return m
}

// parseRecurringPeriod tries to deduce a dt.YearMonthDay from stripe.PriceRecurring
// when it is not set in stripe's Metadata field
func parseRecurringPeriod(r *stripe.PriceRecurring) dt.YearMonthDay {
	// When PriceRecurring is nil, we know nothing
	// about the subscription period.
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

// ToParams outputs to the format needed by Stripe SDK.
func (m StripePriceMeta) ToParams() map[string]string {
	// Format time into ISO8601 set in UTC.
	start := ""
	end := ""
	if !m.StartUTC.IsZero() {
		start = m.StartUTC.In(time.UTC).Format(time.RFC3339)
	}

	if !m.EndUTC.IsZero() {
		end = m.EndUTC.In(time.UTC).Format(time.RFC3339)
	}

	return map[string]string{
		"tier":         m.Tier.String(),
		"years":        strconv.FormatInt(m.PeriodCount.Years, 10),
		"months":       strconv.FormatInt(m.PeriodCount.Months, 10),
		"days":         strconv.FormatInt(m.PeriodCount.Days, 10),
		"introductory": strconv.FormatBool(m.Introductory),
		"start_utc":    start,
		"end_utc":      end,
	}
}

// StripePriceMetaFromFtc extracts metadata from
// an ftc equivalent price.
// TODO: this might no longer be needed.
// Deprecated.
func StripePriceMetaFromFtc(p FtcPrice) StripePriceMeta {
	return StripePriceMeta{
		Introductory: p.IsOneTime(),
		PeriodCount:  p.PeriodCount.YearMonthDay,
		Tier:         p.Tier,
		StartUTC:     p.StartUTC,
		EndUTC:       p.EndUTC,
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
	start := m["start_utc"]
	end := m["end_utc"]

	var startTime, endTime time.Time
	if start != "" {
		startTime, _ = time.Parse(time.RFC3339, start)
	}

	if end != "" {
		endTime, _ = time.Parse(time.RFC3339, end)
	}

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
