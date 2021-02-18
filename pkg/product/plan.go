package product

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
)

type DailyCost struct {
	Holder   string
	Replacer string
}

func NewDailyCostOfYear(price float64) DailyCost {
	return DailyCost{
		Holder:   "{{dailyAverageOfYear}}",
		Replacer: FormatMoney(price / 360),
	}
}

func NewDailyCostOfMonth(price float64) DailyCost {
	return DailyCost{
		Holder:   "{{dailyAverageOfMonth}}",
		Replacer: FormatMoney(price / 30),
	}
}

type Plan struct {
	ID        string  `json:"id" db:"plan_id"`
	ProductID string  `json:"productId" db:"product_id"`
	Price     float64 `json:"price" db:"price"`
	Edition
	Description string `json:"description" db:"description"`
}

// PaymentTitle is used as the value of `subject` for alipay,
// and `body` for wechat pay.
// * 订阅FT中文网标准会员/年
// * 续订FT中文网标准会员/年
// * 升级订阅FT中文网高端会员/年
func (p Plan) PaymentTitle(k enum.OrderKind) string {

	return fmt.Sprintf("%sFT中文网%s", k.StringSC(), p.Edition.StringCN())
}

type ExpandedPlan struct {
	Plan
	Discount Discount `json:"discount"`
}

// ExpandedPlanSchema contains a plans and its discount.
type ExpandedPlanSchema struct {
	PlanID    string  `db:"plan_id"`
	ProductID string  `db:"product_id"`
	PlanPrice float64 `db:"price"`
	Edition
	PlanDesc string `db:"description"`
	Discount
}

// GroupPricesOfProduct groups plans by product id.
func GroupPricesOfProduct(prices []Price) map[string][]Price {
	var g = make(map[string][]Price)

	for _, v := range prices {
		found, ok := g[v.ProductID]
		if ok {
			found = append(found, v)
		} else {
			found = []Price{v}
		}
		g[v.ProductID] = found
	}

	return g
}
