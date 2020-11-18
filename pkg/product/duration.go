package product

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"time"
)

var cycleDays = map[enum.Cycle]int64{
	enum.CycleYear:  365,
	enum.CycleMonth: 30,
}

// Duration is the number of billing cycles and trial days a subscription purchased.
type Duration struct {
	CycleCount int64 `json:"cycleCount" db:"cycle_count"`
	ExtraDays  int64 `json:"extraDays" db:"extra_days"`
}

// DefaultDuration specifies the default value for a duration
func DefaultDuration() Duration {
	return Duration{
		CycleCount: 1,
		ExtraDays:  1,
	}
}

type DurationBuilder struct {
	cycle enum.Cycle
	Duration
}

func NewDurationBuilder(c enum.Cycle, d Duration) DurationBuilder {
	return DurationBuilder{
		cycle:    c,
		Duration: d,
	}
}

func (b DurationBuilder) ToDays() int64 {
	return b.CycleCount*cycleDays[b.cycle] + b.ExtraDays
}

func (b DurationBuilder) ToDateRange(start chrono.Date) (dt.DateRange, error) {
	var endTime time.Time

	switch b.cycle {
	case enum.CycleYear:
		endTime = start.AddDate(int(b.CycleCount), 0, int(b.ExtraDays))

	case enum.CycleMonth:
		endTime = start.AddDate(0, int(b.CycleCount), int(b.ExtraDays))

	default:
		return dt.DateRange{}, errors.New("to date range: invalid billing cycle")
	}

	return dt.DateRange{
		StartDate: start,
		EndDate:   chrono.DateFrom(endTime),
	}, nil
}
