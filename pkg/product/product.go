package product

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"strings"
)

// Product defines a product without plans.
type Product struct {
	ID          string      `json:"id" db:"product_id"`
	Tier        enum.Tier   `json:"tier" db:"tier"`
	Heading     string      `json:"heading" db:"heading"`
	Description null.String `json:"description" db:"description"`
	SmallPrint  null.String `json:"smallPrint" db:"small_print"`
}

// ExpandedProduct contains Product and its plans.
type ExpandedProduct struct {
	Product
	Plans []ExpandedPlan `json:"plans"`
}

// BuildPaywallProducts zips product with its plans.
func BuildPaywallProducts(prods []Product, plans []ExpandedPlan) []ExpandedProduct {
	groupedPlans := GroupPlans(plans)

	var result = make([]ExpandedProduct, 0)

	for _, prod := range prods {
		gPlans, ok := groupedPlans[prod.ID]

		if !ok {
			gPlans = []ExpandedPlan{}
		}

		for _, plan := range gPlans {
			dailyCost := plan.DailyCost()
			desc := strings.Replace(prod.Description.String, dailyCost.Holder, dailyCost.Replacer, 1)
			prod.Description = null.NewString(desc, desc != "")
		}

		result = append(result, ExpandedProduct{
			Product: prod,
			Plans:   gPlans,
		})
	}

	return result
}
