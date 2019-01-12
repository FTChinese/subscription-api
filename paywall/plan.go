package paywall

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"gitlab.com/ftchinese/subscription-api/enum"
)

// Plan is a pricing plan.
type Plan struct {
	Tier        enum.Tier  `json:"tier"`
	Cycle       enum.Cycle `json:"cycle"`
	Price       float64    `json:"price"`
	ID          int        `json:"id"` // 10 for standard and 100 for premium
	Description string     `json:"description"`
	Ignore      bool       `json:"ignore,omitempty"`
}

// PriceForWx calculates price in cent to be used for Wechat pay.
func (p Plan) PriceForWx() int64 {
	return int64(p.Price * 100)
}

// PriceForAli formats price for alipay
func (p Plan) PriceForAli() string {
	return strconv.FormatFloat(p.Price, 'f', 2, 32)
}

// OrderID generates an FT order id based
// on the plan's id, a random number between 100 to 999,
// and unix timestamp.
func (p Plan) OrderID() string {
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between [100, 999)
	rn := 100 + rand.Intn(999-100+1)

	return fmt.Sprintf("FT%03d%d%d", p.ID, rn, time.Now().Unix())
}
