package pw

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"strings"
)

// ProductBody defines a price without plans.
type ProductBody struct {
	ID          string      `json:"id" db:"product_id"`
	Tier        enum.Tier   `json:"tier" db:"tier"`
	Heading     string      `json:"heading" db:"heading"`
	Description null.String `json:"description" db:"description"`
	SmallPrint  null.String `json:"smallPrint" db:"small_print"`
}

// Product describes the data used to present to user on paywall.
type Product struct {
	ProductBody
	Prices []FtcPrice `json:"prices"`
}

// NewPaywallProducts zips price body with its prices.
// Currently we have two ProductBody, and three FtcPrice.
func NewPaywallProducts(prods []ProductBody, prices []FtcPrice) []Product {
	groupedPrices := groupProductPrices(prices)

	var result = make([]Product, 0)

	for _, prod := range prods {
		// Get all prices belong to this price.
		prodPrices, ok := groupedPrices[prod.ID]

		// If nothing found, assign it an empty array.
		if !ok {
			prodPrices = []FtcPrice{}
		}

		// Calculate daily price.
		for _, pp := range prodPrices {
			dailyCost := pp.Original.DailyCost()
			// Replace daily price placeholder in price body.
			desc := strings.Replace(prod.Description.String, dailyCost.Holder, dailyCost.Replacer, 1)
			prod.Description = null.NewString(desc, desc != "")
		}

		result = append(result, Product{
			ProductBody: prod,
			Prices:      prodPrices,
		})
	}

	return result
}

// GroupProductPrices put prices with the same price id into the same group
func groupProductPrices(prices []FtcPrice) map[string][]FtcPrice {
	var g = make(map[string][]FtcPrice)

	for _, p := range prices {
		found, ok := g[p.Original.ProductID]
		if ok {
			found = append(found, p)
		} else {
			found = []FtcPrice{p}
		}
		// Put price of the same price into the same group.
		g[p.Original.ProductID] = found
	}

	return g
}
