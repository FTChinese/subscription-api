package subs

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"time"
)

type PurchasedPeriod struct {
	// Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	StartDate chrono.Date `json:"startDate" db:"start_date"`
	// Membership end date for this order. Depends on start date.
	EndDate chrono.Date `json:"endDate" db:"end_date"`
}

type PeriodBuilder struct {
	product.Edition
	product.Duration
}

func NewPeriodBuilder(e product.Edition, l product.Duration) PeriodBuilder {
	return PeriodBuilder{
		Edition:  e,
		Duration: l,
	}
}

func (b PeriodBuilder) Build(start chrono.Date) (PurchasedPeriod, error) {
	var endTime time.Time

	switch b.Cycle {
	case enum.CycleYear:
		endTime = start.AddDate(int(b.CycleCount), 0, int(b.ExtraDays))

	case enum.CycleMonth:
		endTime = start.AddDate(0, int(b.CycleCount), int(b.ExtraDays))

	default:
		return PurchasedPeriod{}, errors.New("invalid billing cycle")
	}

	return PurchasedPeriod{
		StartDate: start,
		EndDate:   chrono.DateFrom(endTime),
	}, nil
}
