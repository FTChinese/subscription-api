package plan

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"math"
	"strings"
	"time"
)

// BasePlan includes a product and a plan billing cycle to identify the plan to subscribe.
type BasePlan struct {
	Tier       enum.Tier  `json:"tier" db:"sub_tier"`
	Cycle      enum.Cycle `json:"cycle" db:"sub_cycle"`
	LegacyTier null.Int   `json:"-" db:"vip_type"`
}

// NamedKey create a unique name for the point in the plane.
func (c BasePlan) NamedKey() string {
	return c.Tier.String() + "_" + c.Cycle.String()
}

// Plan is a pricing plan.
// The list price is the price that buyers pay for your product or service without any discounts.
// The net price of a product or service is the actual price that customers pay for the product or service.
type Plan struct {
	BasePlan
	ListPrice float64 `json:"listPrice" db:"price"`
	NetPrice  float64 `json:"netPrice" db:"amount"`
	Duration
	Currency         string `json:"currency" db:"currency"`
	Title            string `json:"description" db:"title"`
	stripeLivePlanID string `json:"-"`
	stripeTestPlanID string `json:"-"`
	AppleProductID   string `json:"-"`
}

func (p Plan) GetStripePlanID(live bool) string {
	if live {
		return p.stripeLivePlanID
	}

	return p.stripeTestPlanID
}

// withSandboxPrice returns the sandbox version of a plan.
func (p Plan) withSandboxPrice() Plan {
	p.NetPrice = 0.01
	return p
}

func (p Plan) WithStripePrice(sp stripe.Plan) Plan {
	p.ListPrice = float64(sp.Amount / 100)
	p.NetPrice = p.ListPrice
	p.Currency = string(sp.Currency)

	return p
}

// WithUpgrade creates an upgrading plan.
func (p Plan) WithUpgrade(balance float64) Plan {

	if balance < p.NetPrice {
		p.NetPrice = p.NetPrice - balance
	} else {
		p.NetPrice = 0
	}

	q := p.CalculateConversion(balance)

	p.CycleCount = q.CycleCount
	p.ExtraDays = q.ExtraDays
	p.Title = "FT中文网 - 升级高端会员"

	return p
}

// ConvertBalance checks to see how many cycles and extra
// days a user's balance could be exchanged.
func (p Plan) CalculateConversion(balance float64) Duration {

	if balance <= p.NetPrice {
		return Duration{
			CycleCount: 1,
			ExtraDays:  1,
		}
	}

	cycles, r := util.Division(balance, p.NetPrice)

	days := math.Ceil(r * 365 / p.NetPrice)

	return Duration{
		CycleCount: cycles,
		ExtraDays:  int64(days),
	}
}

func (p Plan) GetPeriodEnd(start time.Time) (time.Time, error) {

	switch p.Cycle {
	case enum.CycleYear:
		return start.AddDate(int(p.CycleCount), 0, int(p.ExtraDays)), nil

	case enum.CycleMonth:
		return start.AddDate(0, int(p.CycleCount), int(p.ExtraDays)), nil

	default:
		return time.Time{}, errors.New("invalid plan cycle")
	}
}

// Desc is used for displaying to user.
// The price show here is not the final price user paid.
func (p Plan) Desc() string {
	return fmt.Sprintf(
		"%s %s%.2f/%s",
		p.Tier.StringCN(),
		strings.ToUpper(p.Currency),
		p.ListPrice,
		p.Cycle.StringCN(),
	)
}
