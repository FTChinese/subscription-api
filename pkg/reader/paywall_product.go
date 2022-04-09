package reader

// PaywallProduct describes the data used to present to user on paywall.
type PaywallProduct struct {
	Product
	Prices []PaywallPrice `json:"prices"`
}

// NewPaywallProducts zips price body with its prices.
// Currently, we have two Product, and three price.PaywallPrice.
func NewPaywallProducts(prods []Product, prices []PaywallPrice) []PaywallProduct {
	groupedPrices := groupProductPrices(prices)

	var result = make([]PaywallProduct, 0)

	for _, prod := range prods {
		// Get all prices belong to this price.
		prodPrices, ok := groupedPrices[prod.ID]

		// If nothing found, assign it an empty array.
		if !ok {
			prodPrices = []PaywallPrice{}
		}

		result = append(result, PaywallProduct{
			Product: prod,
			Prices:  prodPrices,
		})
	}

	return result
}

// GroupProductPrices put prices with the same price id into the same group
func groupProductPrices(prices []PaywallPrice) map[string][]PaywallPrice {
	var g = make(map[string][]PaywallPrice)

	for _, p := range prices {
		found, ok := g[p.ProductID]
		if ok {
			found = append(found, p)
		} else {
			found = []PaywallPrice{p}
		}
		// Put price of the same price into the same group.
		g[p.ProductID] = found
	}

	return g
}
