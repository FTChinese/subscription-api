package paywall

import (
	"time"

	"github.com/FTChinese/go-rest/chrono"
)

// Promotion is a scheduled promotion event.
type Promotion struct {
	StartUTC  chrono.Time `json:"startAt"`
	EndUTC    chrono.Time `json:"endAt"`
	Banner    Banner      `json:"banner"`
	Pricing   Pricing     `json:"plans"`
	CreatedAt chrono.Time `json:"createdAt"`
}

// IsInEffect tests if now falls within the range of
// a promotion's start and end time.
func (p Promotion) IsInEffect() bool {
	now := time.Now()

	// Start <------ now -------> End
	if p.StartUTC.After(now) || p.EndUTC.Before(now) {
		return false
	}
	return true
}
