package dt

import "github.com/FTChinese/go-rest/enum"

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

// NewYearMonthDay creates a new instance for a enum.Cycle.
func NewYearMonthDay(cycle enum.Cycle) YearMonthDay {
	switch cycle {
	case enum.CycleYear:
		return YearMonthDay{
			Years:  1,
			Months: 0,
			Days:   1,
		}

	case enum.CycleMonth:
		return YearMonthDay{
			Years:  0,
			Months: 1,
			Days:   1,
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
			Days:   int64(n),
		}

	case enum.CycleMonth:
		return YearMonthDay{
			Years:  0,
			Months: int64(n),
			Days:   int64(n),
		}

	default:
		return YearMonthDay{}
	}
}

// TotalDays calculates the number of days of by adding the days of the year, month and days.
func (y YearMonthDay) TotalDays() int64 {
	return y.Years*daysOfYear + y.Months*daysOfMonth + y.Days
}

// Add adds two instances.
func (y YearMonthDay) Add(other YearMonthDay) YearMonthDay {
	y.Years = y.Years + other.Years
	y.Months = y.Months + other.Months
	y.Days = y.Days + other.Days

	return y
}
