package paywall

import (
	"strconv"

	"github.com/FTChinese/go-rest/enum"
)

// Plan is a pricing plan.
// The list price is the price that buyers pay for your product or service without any discounts.
// The net price of a product or service is the actual price that customers pay for the product or service.
type Plan struct {
	Tier        enum.Tier  `json:"tier"`
	Cycle       enum.Cycle `json:"cycle"`
	ListPrice   float64    `json:"listPrice"`
	NetPrice    float64    `json:"netPrice"`
	Description string     `json:"description"`
	Ignore      bool       `json:"ignore,omitempty"`
}

// WxPrice converts price to Wechat pay format.
func (p Plan) WxPrice() int64 {
	return int64(p.ListPrice * 100)
}

// AliPrice converts price to Alipay format.
func (p Plan) AliPrice() string {
	return strconv.FormatFloat(p.ListPrice, 'f', 2, 32)
}

// OrderID generates an FT order id based
// on the plan's id, a random number between 100 to 999,
// and unix timestamp.
// func (p Plan) OrderID() string {
// 	rand.Seed(time.Now().UnixNano())

// 	// Generate a random number between [100, 999)
// 	rn := 100 + rand.Intn(999-100+1)

// 	return fmt.Sprintf("FT%03d%d%d", p.ID, rn, time.Now().Unix())
// }
