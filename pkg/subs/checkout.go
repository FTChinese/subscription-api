package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
)

// Checkout contains the calculation result of a purchase transaction.
type Checkout struct {
	Kind     enum.OrderKind `json:"kind"`
	Cart     price.Cart     `json:"cart"`
	Payable  price.Charge   `json:"payable"`
	LiveMode bool           `json:"live"`
}

func (c Checkout) WithTest(t bool) Checkout {
	c.LiveMode = !t

	if t {
		c.Payable.Amount = 0.01
	}

	return c
}
