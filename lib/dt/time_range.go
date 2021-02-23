package dt

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"time"
)

type DateTimeRange struct {
	StartUTC chrono.Time `json:"startUtc" db:"start_utc"`
	EndUTC   chrono.Time `json:"endUtc" db:"end_utc"`
}

// DateRange is used to build the subscription period of a one-time purchase.
type DateRange struct {
	// Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	StartDate chrono.Date `json:"startDate" db:"start_date"`
	// Membership end date for this order. Depends on start date.
	EndDate chrono.Date `json:"endDate" db:"end_date"`
}

// NewDateRange creates a new instance of DateRange from a moment, with both starting and ending time set to the same value.
func NewDateRange(start time.Time) DateRange {
	return DateRange{
		StartDate: chrono.DateFrom(start),
		EndDate:   chrono.DateFrom(start),
	}
}

// WithCycle adds a billing cycle to end date.
func (d DateRange) WithCycle(cycle enum.Cycle) DateRange {
	switch cycle {
	case enum.CycleYear:
		d.EndDate = chrono.DateFrom(d.EndDate.AddDate(1, 0, 0))

	case enum.CycleMonth:
		d.EndDate = chrono.DateFrom(d.EndDate.AddDate(0, 1, 0))
	}

	return d
}

// WithCycleN adds n cycles to end date.
func (d DateRange) WithCycleN(cycle enum.Cycle, n int) DateRange {
	switch cycle {
	case enum.CycleYear:
		d.EndDate = chrono.DateFrom(d.EndDate.AddDate(n, 0, 0))
	case enum.CycleMonth:
		d.EndDate = chrono.DateFrom(d.EndDate.AddDate(0, n, 0))
	}

	return d
}

func (d DateRange) AddYears(years int) DateRange {
	d.EndDate = chrono.DateFrom(d.EndDate.AddDate(years, 0, 0))
	return d
}

func (d DateRange) AddMonths(months int) DateRange {
	d.EndDate = chrono.DateFrom(d.EndDate.AddDate(0, months, 0))
	return d
}

func (d DateRange) AddDays(days int) DateRange {
	d.EndDate = chrono.DateFrom(d.EndDate.AddDate(0, 0, days))

	return d
}

// AddDate adds the specified years, months, days to end date.
// This is a simple wrapper of Time.AddDate.
func (d DateRange) AddDate(years, months, days int) DateRange {
	d.EndDate = chrono.DateFrom(d.EndDate.AddDate(years, months, days))

	return d
}
