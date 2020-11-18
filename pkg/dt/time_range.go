package dt

import "github.com/FTChinese/go-rest/chrono"

type DateTimeRange struct {
	StartUTC chrono.Time `json:"startUtc" db:"start_utc"`
	EndUTC   chrono.Time `json:"endUtc" db:"end_utc"`
}

type DateRange struct {
	// Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	StartDate chrono.Date `json:"startDate" db:"start_date"`
	// Membership end date for this order. Depends on start date.
	EndDate chrono.Date `json:"endDate" db:"end_date"`
}
