package dt

import "github.com/FTChinese/go-rest/enum"

const (
	daysOfYear  = 366
	daysOfMonth = 31
)

type YearMonthDay struct {
	Years  int64 `json:"years" db:"years"`
	Months int64 `json:"months" db:"months"`
	Days   int64 `json:"days" db:"days"`
}

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

func (y YearMonthDay) TotalDays() int64 {
	return y.Years*daysOfYear + y.Months*daysOfMonth + y.Days
}
