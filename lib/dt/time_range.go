package dt

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"time"
)

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func NewTimeRange(start time.Time) TimeRange {
	return TimeRange{
		Start: start,
		End:   start,
	}
}

func (r TimeRange) WithDate(d YearMonthDay) TimeRange {
	r.End = r.End.AddDate(int(d.Years), int(d.Months), int(d.Days))

	return r
}

func (r TimeRange) WithCycle(cycle enum.Cycle) TimeRange {
	switch cycle {
	case enum.CycleYear:
		r.End = r.End.AddDate(1, 0, 0)

	case enum.CycleMonth:
		r.End = r.End.AddDate(0, 1, 0)
	}

	return r
}

// WithCycleN adds n cycles to end date.
func (r TimeRange) WithCycleN(cycle enum.Cycle, n int) TimeRange {
	switch cycle {
	case enum.CycleYear:
		r.End = r.End.AddDate(n, 0, 0)
	case enum.CycleMonth:
		r.End = r.End.AddDate(0, n, 0)
	}

	return r
}

func (r TimeRange) AddYears(years int) TimeRange {
	r.End = r.End.AddDate(years, 0, 0)
	return r
}

func (r TimeRange) AddMonths(months int) TimeRange {
	r.End = r.End.AddDate(0, months, 0)
	return r
}

func (r TimeRange) AddDays(days int) TimeRange {
	r.End = r.End.AddDate(0, 0, days)

	return r
}

// AddDate adds the specified years, months, days to end date.
// This is a simple wrapper of Time.AddDate.
func (r TimeRange) AddDate(years, months, days int) TimeRange {
	r.End = r.End.AddDate(years, months, days)

	return r
}

func (r TimeRange) ToDatePeriod() DatePeriod {
	return DatePeriod{
		StartDate: chrono.DateFrom(r.Start),
		EndDate:   chrono.DateFrom(r.End),
	}
}

func (r TimeRange) ToDateTimePeriod() DateTimePeriod {
	return DateTimePeriod{
		StartUTC: chrono.TimeFrom(r.Start),
		EndUTC:   chrono.TimeFrom(r.End),
	}
}

type DateTimePeriod struct {
	StartUTC chrono.Time `json:"startUtc" db:"start_utc"`
	EndUTC   chrono.Time `json:"endUtc" db:"end_utc"`
}

func (p DateTimePeriod) ToDatePeriod() DatePeriod {
	return DatePeriod{
		StartDate: chrono.DateFrom(p.StartUTC.Time),
		EndDate:   chrono.DateFrom(p.EndUTC.Time),
	}
}

// DatePeriod is used to build the subscription period of a one-time purchase.
type DatePeriod struct {
	// Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	StartDate chrono.Date `json:"startDate" db:"start_date"`
	// Membership end date for this order. Depends on start date.
	EndDate chrono.Date `json:"endDate" db:"end_date"`
}
