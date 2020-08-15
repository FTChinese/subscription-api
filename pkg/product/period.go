package product

import "github.com/FTChinese/go-rest/chrono"

type Period struct {
	StartUTC chrono.Time `json:"startUtc" db:"start_utc"`
	EndUTC   chrono.Time `json:"endUtc" db:"end_utc"`
}
