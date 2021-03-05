package dt

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
)

type PurchasedTimeRangeBuilder struct {
	ConfirmedAt    chrono.Time
	ExpirationDate chrono.Date
	Date           YearMonthDay
	OrderKind      enum.OrderKind
}

func (b PurchasedTimeRangeBuilder) Build() (TimeRange, error) {
	switch b.OrderKind {

	case enum.OrderKindCreate, enum.OrderKindRenew:
		startTime := PickLater(b.ConfirmedAt.Time, b.ExpirationDate.Time)
		return NewTimeRange(startTime).
			WithDate(b.Date), nil

	case enum.OrderKindUpgrade:
		return NewTimeRange(b.ConfirmedAt.Time).
			WithDate(b.Date), nil

	case enum.OrderKindAddOn:
		return TimeRange{}, nil
	}

	return TimeRange{}, errors.New("cannot determine purchased time range due to unknown order kind")
}
