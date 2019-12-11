package plan

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/stripe/stripe-go"
	"strings"
)

// BasePlan includes a product and a plan billing cycle to identify the plan to subscribe.
type BasePlan struct {
	Tier  enum.Tier  `json:"tier" db:"plan_tier"`
	Cycle enum.Cycle `json:"cycle" db:"plan_cycle"`
}

func (p BasePlan) IsZero() bool {
	return p.Tier == enum.TierNull && p.Cycle == enum.CycleNull
}

// NamedKey create a unique name for the point in the plane.
func (p BasePlan) NamedKey() string {
	return p.Tier.String() + "_" + p.Cycle.String()
}

// Desc produces a human readable string of this plan.
// * 标准会员/年
// * 标准会员/月
// * 高端会员/年
func (p BasePlan) Desc() string {
	return p.Tier.StringCN() + "/" + p.Cycle.StringCN()
}

// Plan is a pricing plan.
// The list price is the price that buyers pay for your product or service without any discounts.
// The net price of a product or service is the actual price that customers pay for the product or service.
type Plan struct {
	BasePlan
	Price            float64 `json:"price" db:"plan_price"`       // Price of a plan, prior to discount.
	Amount           float64 `json:"amount" db:"plan_amount"`     // Actual price paid.
	Currency         string  `json:"currency" db:"plan_currency"` // in which currency.
	Title            string  `json:"description"`
	stripeLivePlanID string  `json:"-"`
	stripeTestPlanID string  `json:"-"`
	AppleProductID   string  `json:"-"`

	ListPrice float64 `json:"listPrice"` // Deprecated
	NetPrice  float64 `json:"netPrice"`  // Deprecated
}

// GetTitle compose the message shown for wxpay or alipay.
// * 订阅FT中文网标准会员/年
// * 续订FT中文网标准会员/年
// * 升级订阅FT中文网高端会员/年
func (p Plan) GetTitle(k SubsKind) string {
	return fmt.Sprintf("%sFT中文网%s/%s", k.StringCN(), p.Tier.StringCN(), p.Cycle.StringCN())
}

// Desc is used for displaying to user.
// The price show here is not the final price user paid.
// 标准会员/年 CNY258.00
func (p Plan) Desc() string {
	return fmt.Sprintf(
		"%s %s%.2f",
		p.BasePlan.Desc(),
		strings.ToUpper(p.Currency),
		p.Price,
	)
}

func (p Plan) GetStripePlanID(live bool) string {
	if live {
		return p.stripeLivePlanID
	}

	return p.stripeTestPlanID
}

func (p Plan) WithStripePrice(sp stripe.Plan) Plan {

	p.Price = float64(sp.Amount / 100)
	p.Amount = p.Price
	p.Currency = string(sp.Currency)

	p.ListPrice = p.Price
	p.NetPrice = p.Price

	return p
}
