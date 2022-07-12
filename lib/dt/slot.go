package dt

import (
	"github.com/FTChinese/go-rest/chrono"
	"time"
)

type TimeSlot struct {
	StartUTC chrono.Time `json:"startUtc" db:"start_utc"`
	EndUTC   chrono.Time `json:"endUtc" db:"end_utc"`
}

// Include tests if the specified moment falls in this time slot.
func (ts TimeSlot) Include(t time.Time) bool {
	if t.Before(ts.StartUTC.Time) || t.After(ts.EndUTC.Time) {
		return false
	}

	return true
}

// NowIn tests if current moment falls in this time slot.
func (ts TimeSlot) NowIn() bool {
	return ts.Include(time.Now())
}

type DateSlot struct {
	StartDate chrono.Date `json:"startDate" db:"start_date"`
	EndDate   chrono.Date `json:"endDate" db:"end_date"`
}
