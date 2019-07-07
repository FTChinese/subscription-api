package billing

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
)

type Product struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	CreatedAt chrono.Time `json:"createdAt"`
	UpdatedAt chrono.Time `json:"updatedAt"`
}

type Plan struct {
	ID            string      `json:"id"`
	Active        bool        `json:"active"`
	Currency      string      `json:"currency"`
	Interval      enum.Cycle  `json:"interval"`
	IntervalCount int64       `json:"intervalCount"`
	TrialDays     int64       `json:"trialDays"`
	ProductID     string      `json:"productId"`
	CreatedAt     chrono.Time `json:"createdAt"`
}
