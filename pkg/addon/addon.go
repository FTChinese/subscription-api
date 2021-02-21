package addon

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

var cycleDays = map[enum.Cycle]int64{
	enum.CycleYear:  366,
	enum.CycleMonth: 31,
}

type AddOn struct {
	ID string `json:"id" db:"id"`
	price.Edition
	CycleCount    int64          `json:"cycleCount" db:"cycle_count"`
	DaysRemained  int64          `json:"daysRemained" db:"days_remained"`
	IsCarryOver   bool           `json:"isCarryOver" db:"is_carry_over"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	CompoundID    string         `json:"compoundId" db:"compound_id"`
	OrderID       null.String    `json:"orderId" db:"order_id"`
	PlanID        null.String    `json:"planId" db:"plan_id"`
	CreatedUTC    chrono.Time    `json:"createdUtc" db:"created_utc"`
	ConsumedUTC   chrono.Time    `json:"consumedUtc" db:"consumed_utc"`
}

func (a AddOn) IsZero() bool {
	return a.ID == ""
}

func (a AddOn) GetDays() int64 {
	return a.CycleCount*cycleDays[a.Cycle] + a.DaysRemained
}

// ToReservedDays calculates how many days this add-on could be converted to reserved part of membership.
func (a AddOn) ToReservedDays() ReservedDays {
	switch a.Tier {
	case enum.TierStandard:
		return ReservedDays{
			Standard: a.GetDays(),
			Premium:  0,
		}
	case enum.TierPremium:
		return ReservedDays{
			Standard: 0,
			Premium:  a.GetDays(),
		}

	default:
		return ReservedDays{}
	}
}

func (a AddOn) WithOrderID(id string) AddOn {
	a.OrderID = null.StringFrom(id)
	return a
}
