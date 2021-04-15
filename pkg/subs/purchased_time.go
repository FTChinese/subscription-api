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
	Date           dt.YearMonthDay // Purchased period.
	OrderKind      enum.OrderKind
}

func (b PurchasedTimeParams) Build() (dt.TimeRange, error) {
	switch b.OrderKind {

	case enum.OrderKindCreate, enum.OrderKindRenew:
		startTime := dt.PickLater(b.ConfirmedAt.Time, b.ExpirationDate.Time)
		return dt.NewTimeRange(startTime).
			WithDate(b.Date), nil

	case enum.OrderKindUpgrade:
		return dt.NewTimeRange(b.ConfirmedAt.Time).
			WithDate(b.Date), nil

	case enum.OrderKindAddOn:
		return dt.TimeRange{}, nil
	}

	return dt.TimeRange{}, errors.New("cannot determine purchased time range due to unknown order kind")
}
