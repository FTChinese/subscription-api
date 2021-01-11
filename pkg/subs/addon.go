package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

var cycleDays = map[enum.Cycle]int64{
	enum.CycleYear:  366,
	enum.CycleMonth: 31,
}

type AddOn struct {
	ID string `json:"id" db:"id"`
	product.Edition
	CycleCount    int64          `json:"cycleCount" db:"cycle_count"`
	DaysRemained  int64          `json:"daysRemained" db:"days_remained"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	OrderID       null.String    `json:"orderId" db:"order_id"`
	CompoundID    string         `json:"compoundId" db:"compound_id"`
	CreatedUTC    chrono.Time    `json:"createdUtc" db:"created_utc"`
	ConsumedUTC   chrono.Time    `json:"consumedUtc" db:"consumed_utc"`
}

func NewAddOn(o Order) AddOn {
	return AddOn{
		ID:            db.AddOnID(),
		Edition:       o.Edition,
		CycleCount:    1,
		DaysRemained:  trialDays,
		PaymentMethod: o.PaymentMethod,
		OrderID:       null.StringFrom(o.ID),
		CompoundID:    o.CompoundID,
		CreatedUTC:    chrono.TimeNow(),
		ConsumedUTC:   chrono.Time{},
	}
}

// NewUpgradeAddOn moves the remaining days of a standard subscription
// to addon portion upon upgrading to premium.
func NewUpgradeAddOn(o Order, m reader.Membership) AddOn {
	return AddOn{
		ID:            db.AddOnID(),
		Edition:       m.Edition,
		CycleCount:    0,
		DaysRemained:  m.RemainingDays(),
		PaymentMethod: m.PaymentMethod,
		OrderID:       null.StringFrom(o.ID), // Which order caused the current membership to move remaining days to reserved state.
		CompoundID:    o.CompoundID,
		CreatedUTC:    chrono.TimeNow(),
		ConsumedUTC:   chrono.Time{},
	}
}

func (a AddOn) IsZero() bool {
	return a.ID == ""
}

func (a AddOn) GetDays() int64 {
	return a.CycleCount*cycleDays[a.Cycle] + a.DaysRemained
}

func (a AddOn) ToReservedDays() reader.ReservedDays {
	switch a.Tier {
	case enum.TierStandard:
		return reader.ReservedDays{
			Standard: a.GetDays(),
			Premium:  0,
		}
	case enum.TierPremium:
		return reader.ReservedDays{
			Standard: 0,
			Premium:  a.GetDays(),
		}

	default:
		return reader.ReservedDays{}
	}
}
