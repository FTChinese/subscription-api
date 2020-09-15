package subs

import "github.com/FTChinese/subscription-api/pkg/product"

// Checkout is the calculation result of a purchased product.
type Checkout struct {
	product.Charge
	product.Duration
}

func (c Checkout) WithTest() Checkout {
	c.Amount = 0.01

	return c
}
