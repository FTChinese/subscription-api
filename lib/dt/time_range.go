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

func (r TimeRange) WithPeriod(d YearMonthDay) TimeRange {
	r.End = r.End.AddDate(int(d.Years), int(d.Months), int(d.Days))

	return r
}

func (r TimeRange) WithCycle(cycle enum.Cycle) TimeRange {
	return r.WithPeriod(NewYearMonthDay(cycle))
}

// WithCycleN adds n cycles to end date.
func (r TimeRange) WithCycleN(cycle enum.Cycle, n int) TimeRange {
	return r.WithPeriod(NewYearMonthDayN(cycle, n))
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

func (r TimeRange) StartTime() chrono.Time {
	return chrono.TimeFrom(r.Start)
}

func (r TimeRange) EndTime() chrono.Time {
	return chrono.TimeFrom(r.End)
}

func (r TimeRange) StartDate() chrono.Date {
	return chrono.DateFrom(r.Start)
}

func (r TimeRange) EndDate() chrono.Date {
	return chrono.DateFrom(r.End)
}

type ChronoPeriod struct {
	StartUTC chrono.Time `json:"startUtc" db:"start_utc"`
	EndUTC   chrono.Time `json:"endUtc" db:"end_utc"`
}

func (p ChronoPeriod) StartDate() chrono.Date {
	return chrono.DateFrom(p.StartUTC.Time)
}

func (p ChronoPeriod) EndDate() chrono.Date {
	return chrono.DateFrom(p.EndUTC.Time)
}
