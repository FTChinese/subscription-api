package pw

import "github.com/FTChinese/subscription-api/pkg/price"

// PaywallProduct describes the data used to present to user on paywall.
type PaywallProduct struct {
	Product
	Prices []price.PaywallPrice `json:"prices"`
}

// NewPaywallProducts zips price body with its prices.
// Currently, we have two Product, and three price.PaywallPrice.
func NewPaywallProducts(prods []Product, prices []price.PaywallPrice) []PaywallProduct {
	groupedPrices := groupProductPrices(prices)

	var result = make([]PaywallProduct, 0)

	for _, prod := range prods {
		// Get all prices belong to this price.
		prodPrices, ok := groupedPrices[prod.ID]

		// If nothing found, assign it an empty array.
		if !ok {
			prodPrices = []price.PaywallPrice{}
		}

		// Calculate daily price.
		//for _, pp := range prodPrices {
		//	dailyCost := pp.DailyCost()
		//	// Replace daily price placeholder in price body.
		//	desc := strings.Replace(prod.Description.String, dailyCost.Holder, dailyCost.Replacer, 1)
		//	prod.Description = null.NewString(desc, desc != "")
		//}

		result = append(result, PaywallProduct{
			Product: prod,
			Prices:  prodPrices,
		})
	}

	return result
}

// GroupProductPrices put prices with the same price id into the same group
func groupProductPrices(prices []price.PaywallPrice) map[string][]price.PaywallPrice {
	var g = make(map[string][]price.PaywallPrice)

	for _, p := range prices {
		found, ok := g[p.ProductID]
		if ok {
			found = append(found, p)
		} else {
			found = []price.PaywallPrice{p}
		}
		// Put price of the same price into the same group.
		g[p.ProductID] = found
	}

	return g
}
