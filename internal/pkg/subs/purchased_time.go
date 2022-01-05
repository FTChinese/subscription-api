package subs

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
)

// PurchasedTimeParams is used to deduce an order's purchase period.
type PurchasedTimeParams struct {
	ConfirmedAt    chrono.Time     // When the order is confirmed
	ExpirationDate chrono.Date     // When the current membership will expire.
	PeriodCount    dt.YearMonthDay // Purchased period.
	OrderKind      enum.OrderKind
}

// Build determines an order's purchased time range.
// For order kid of create or renew, always pick the latest time from confirmation time
// and current membership's expiration time.
// For addon, time range if unknown until a future moment.
func (b PurchasedTimeParams) Build() (dt.TimeRange, error) {
	switch b.OrderKind {

	case enum.OrderKindCreate, enum.OrderKindRenew:
		startTime := dt.PickLater(b.ConfirmedAt.Time, b.ExpirationDate.Time)
		return dt.NewTimeRange(startTime).
			WithPeriod(b.PeriodCount), nil

	// For upgrade, it always starts immediately.
	case enum.OrderKindUpgrade:
		return dt.NewTimeRange(b.ConfirmedAt.Time).
			WithPeriod(b.PeriodCount), nil

	// For addon, you should not extend current subscription
	// period.
	case enum.OrderKindAddOn:
		return dt.TimeRange{}, nil
	}

	return dt.TimeRange{}, errors.New("cannot determine purchased time range due to unknown order kind")
}
