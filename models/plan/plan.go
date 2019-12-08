package plan

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/stripe/stripe-go"
	"strings"
)

// BasePlan includes a product and a plan billing cycle to identify the plan to subscribe.
type BasePlan struct {
	Tier  enum.Tier  `json:"tier" db:"sub_tier"`
	Cycle enum.Cycle `json:"cycle" db:"sub_cycle"`
}

func (p BasePlan) IsZero() bool {
	return p.Tier == enum.TierNull && p.Cycle == enum.CycleNull
}

// NamedKey create a unique name for the point in the plane.
func (p BasePlan) NamedKey() string {
	return p.Tier.String() + "_" + p.Cycle.String()
}

// Plan is a pricing plan.
// The list price is the price that buyers pay for your product or service without any discounts.
// The net price of a product or service is the actual price that customers pay for the product or service.
type Plan struct {
	BasePlan
	ListPrice float64 `json:"listPrice"`          // Deprecate
	NetPrice  float64 `json:"netPrice"`           // Deprecate
	Price     float64 `json:"price" db:"price"`   // Price of a plan, prior to discount.
	Amount    float64 `json:"amount" db:"amount"` // Actually paid amount.
	//Duration                 // This should be removed.
	Currency         string `json:"currency" db:"currency"`
	Title            string `json:"description"`
	stripeLivePlanID string `json:"-"`
	stripeTestPlanID string `json:"-"`
	AppleProductID   string `json:"-"`
}

// GetTitle compose the message shown for wxpay or alipay.
// * 订阅FT中文网标准会员/年
// * 续订FT中文网标准会员/年
// * 升级订阅FT中文网高端会员/年
func (p Plan) GetTitle(k SubsKind) string {
	return fmt.Sprintf("%sFT中文网%s/%s", k.StringCN(), p.Tier.StringCN(), p.Cycle.StringCN())
}

func (p Plan) GetStripePlanID(live bool) string {
	if live {
		return p.stripeLivePlanID
	}

	return p.stripeTestPlanID
}

// withSandboxPrice returns the sandbox version of a plan.
//func (p Plan) withSandboxPrice() Plan {
//	p.NetPrice = 0.01
//	p.Amount = 0.01
//	return p
//}

func (p Plan) WithStripePrice(sp stripe.Plan) Plan {
	p.ListPrice = float64(sp.Amount / 100)
	p.NetPrice = p.ListPrice
	p.Price = p.ListPrice
	p.Amount = p.NetPrice
	p.Currency = string(sp.Currency)

	return p
}

// WithUpgrade creates an upgrading plan.
//func (p Plan) WithUpgrade(balance float64) Plan {
//
//	if balance < p.Price {
//		p.Price = p.Price - balance
//	} else {
//		p.Price = 0
//	}
//
//	p.Amount = p.Price
//
//	dur := p.CalculateConversion(balance)
//
//	//p.CycleCount = dur.CycleCount
//	//p.ExtraDays = dur.ExtraDays
//	p.Title = "FT中文网 - 升级高端会员"
//
//	return p
//}

// ConvertBalance checks to see how many cycles and extra
// days a user's balance could be exchanged.
//func (p Plan) CalculateConversion(balance float64) Duration {
//
//	if balance <= p.Amount {
//		return Duration{
//			CycleCount: 1,
//			ExtraDays:  1,
//		}
//	}
//
//	cycles, r := util.Division(balance, p.NetPrice)
//
//	days := math.Ceil(r * 365 / p.NetPrice)
//
//	return Duration{
//		CycleCount: cycles,
//		ExtraDays:  int64(days),
//	}
//}

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
