package dt

import (
	"github.com/FTChinese/go-rest/enum"
)

const (
	daysOfYear  = 366
	daysOfMonth = 31
)

// YearMonthDay is the unit of a enum.Cycle.
type YearMonthDay struct {
	Years  int64 `json:"years" db:"years"`
	Months int64 `json:"months" db:"months"`
	Days   int64 `json:"days" db:"days"`
}

// NewYearMonthDay creates a new instance for an enum.Cycle.
func NewYearMonthDay(cycle enum.Cycle) YearMonthDay {
	switch cycle {
	case enum.CycleYear:
		return YearMonthDay{
			Years:  1,
			Months: 0,
			Days:   0,
		}

	case enum.CycleMonth:
		return YearMonthDay{
			Years:  0,
			Months: 1,
			Days:   0,
		}

	default:
		return YearMonthDay{}
	}
}

// NewYearMonthDayN creates a new instance for n enum.Cycle.
func NewYearMonthDayN(cycle enum.Cycle, n int) YearMonthDay {
	switch cycle {
	case enum.CycleYear:
		return YearMonthDay{
			Years:  int64(n),
			Months: 0,
			Days:   0,
		}

	case enum.CycleMonth:
		return YearMonthDay{
			Years:  0,
			Months: int64(n),
			Days:   0,
		}

	default:
		return YearMonthDay{}
	}
}

// TotalDays calculates the number of days of by adding the days of the year, month and days.
func (y YearMonthDay) TotalDays() int64 {
	return y.Years*daysOfYear + y.Months*daysOfMonth + y.Days
}

func (y YearMonthDay) EqCycle() enum.Cycle {
	if y.Years > 0 {
		return enum.CycleYear
	}

	if y.Months > 0 {
		return enum.CycleMonth
	}

	return enum.CycleMonth
}

// Plus adds two instances.
func (y YearMonthDay) Plus(other YearMonthDay) YearMonthDay {
	y.Years = y.Years + other.Years
	y.Months = y.Months + other.Months
	y.Days = y.Days + other.Days

	return y
}

func (y YearMonthDay) AddDays(n int64) YearMonthDay {
	y.Days += n

	return y
}

func (y YearMonthDay) AddMonths(n int64) YearMonthDay {
	y.Months += n

	return y
}

func (y YearMonthDay) AddYears(n int64) YearMonthDay {
	y.Years += n

	return y
}

func (y YearMonthDay) IsZero() bool {
	return y.Years == 0 && y.Months == 0 && y.Days == 0
}

// IsSingular checks to see if there's only
// Years or Months field having value set.
// This is mostly useful when you want to format
// a human-readable string.
// If there's only one year or one month set,
// you can format string like:
// - Standard/Month
// - Standard/Year
// otherwise you should explicitly tell user
// the period they are purchasing:
// - Standard/2 years 3 months 7 days
func (y YearMonthDay) IsSingular() bool {
	if y.Years == 1 && y.Months == 0 && y.Days == 0 {
		return true
	}

	if y.Years == 0 && y.Months == 1 && y.Days == 0 {
		return true
	}

	return false
}
