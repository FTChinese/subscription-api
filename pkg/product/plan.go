package product

type Plan struct {
	ID        string  `json:"id" db:"plan_id"`
	ProductID string  `json:"productId" db:"product_id"`
	Price     float64 `json:"price" db:"price"`
	Edition
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
