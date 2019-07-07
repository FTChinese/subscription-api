package paywall

import "github.com/FTChinese/go-rest/enum"

// Plan is a pricing plan.
// The list price is the price that buyers pay for your product or service without any discounts.
// The net price of a product or service is the actual price that customers pay for the product or service.
type Plan struct {
	Tier        enum.Tier  `json:"tier"` // This is product.
	Cycle       enum.Cycle `json:"cycle"`
	ListPrice   float64    `json:"listPrice"`
	NetPrice    float64    `json:"netPrice"`
	Description string     `json:"description"`
}

func (p Plan) ProductID() string {
	return p.Tier.String() + "_" + p.Cycle.String()
}

func (p Plan) StripePrice() int64 {
	return int64(p.NetPrice * 100)
}
