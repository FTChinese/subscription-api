package product

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

type Plan struct {
	ID        string  `json:"id" db:"plan_id"`
	ProductID string  `json:"productId" db:"product_id"`
	Price     float64 `json:"price" db:"price"`
	Edition
	Description null.String `json:"description" db:"description"`
}

// PaymentTitle is used as the value of `subject` for alipay,
// and `body` for wechat pay.
// * 订阅FT中文网标准会员/年
// * 续订FT中文网标准会员/年
// * 升级订阅FT中文网高端会员/年
func (p Plan) PaymentTitle(k enum.OrderKind) string {
	return fmt.Sprintf("%sFT中文网%s", k.StringSC(), p.Edition.StringCN())
}

// IntentPlan describes user's intent plan of subscription.
// This is used to keep backward-compatible before
// we persist plans in db.
type IntentPlan struct {
	Plan
	Charge
}

type ExpandedPlan struct {
	Plan
	Discount Discount `json:"discount"`
}

// Amount calculates how much a user should pay.
func (e ExpandedPlan) Amount() float64 {
	if e.Discount.IsValid() {
		return e.Price - e.Discount.PriceOff.Float64
	}

	return e.Price
}

// ExpandedPlanSchema contains a plans and its discount.
type ExpandedPlanSchema struct {
	Plan
	Discount
}

func (s ExpandedPlanSchema) ExpandedPlan() ExpandedPlan {
	return ExpandedPlan{
		Plan:     s.Plan,
		Discount: s.Discount,
	}
}

// GroupPlans groups plans by product id.
func GroupPlans(plans []ExpandedPlan) map[string][]ExpandedPlan {
	var g = make(map[string][]ExpandedPlan)

	for _, v := range plans {
		found, ok := g[v.ProductID]
		if ok {
			found = append(found, v)
		} else {
			found = []ExpandedPlan{v}
		}
		g[v.ProductID] = found
	}

	return g
}
