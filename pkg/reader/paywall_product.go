package reader

import "github.com/FTChinese/subscription-api/pkg/price"

// PaywallProduct describes the data used to present to user on paywall.
type PaywallProduct struct {
	Product
	Prices []PaywallPrice `json:"prices"` // Deprecated. Hoisted to parent level to flatten data structure.
}

// NewPaywallProducts zips price body with its prices.
// Currently, we have two Product, and three price.PaywallPrice.
// Deprecated.
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

// GroupProductPrices put prices with the same product id into the same group
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

type classifiedPrices struct {
	recurring []PaywallPrice
	intro     PaywallPrice
}

func groupProductPricesV2(prices []PaywallPrice) map[string]classifiedPrices {
	var g = make(map[string]classifiedPrices)

	for _, p := range prices {
		found, ok := g[p.ProductID]
		if p.Kind == price.KindOneTime {
			found.intro = p
		} else {
			if ok {
				found.recurring = append(found.recurring, p)
			} else {
				found.recurring = []PaywallPrice{p}
			}
		}

		// Put price of the same price into the same group.
		g[p.ProductID] = found
	}

	return g
}

// NewPaywallProducts zips price body with its prices.
// Currently, we have two Product, and three price.PaywallPrice.
func NewPaywallProductsV2(prods []Product, prices []PaywallPrice) []PaywallProduct {
	groupedPrices := groupProductPricesV2(prices)

	var result = make([]PaywallProduct, 0)

	for _, prod := range prods {
		// Get all prices belong to this price.
		prodPrices, ok := groupedPrices[prod.ID]

		if !ok {
			// If nothing found, assign it an empty array.
			result = append(result, PaywallProduct{
				Product: prod,
				Prices:  []PaywallPrice{},
			})
		} else {
			prod.Introductory = price.FtcPriceJSON{
				FtcPrice: prodPrices.intro.FtcPrice,
			}
			result = append(result, PaywallProduct{
				Product: prod,
				Prices:  prodPrices.recurring,
			})
		}
	}

	return result
}
