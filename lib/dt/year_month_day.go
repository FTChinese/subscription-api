package dt

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"time"
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

func (y YearMonthDay) EndTime() time.Time {
	return time.Now().AddDate(int(y.Years), int(y.Months), int(y.Days))
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

// Add adds two instances.
func (y YearMonthDay) Add(other YearMonthDay) YearMonthDay {
	y.Years = y.Years + other.Years
	y.Months = y.Months + other.Months
	y.Days = y.Days + other.Days

	return y
}

func (y YearMonthDay) IsZero() bool {
	return y.Years == 0 && y.Months == 0 && y.Days == 0
}

type YearMonthDayJSON struct {
	YearMonthDay
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (j YearMonthDayJSON) Value() (driver.Value, error) {
	if j.IsZero() {
		return nil, nil
	}

	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (j *YearMonthDayJSON) Scan(src interface{}) error {
	if src == nil {
		*j = YearMonthDayJSON{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp YearMonthDayJSON
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*j = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to YearMonthDayJSON")
	}
}
