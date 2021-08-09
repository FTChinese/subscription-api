package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// Checkout contains the calculation result of a purchase transaction.
type Checkout struct {
	Kind     enum.OrderKind
	Price    price.Price
	Offer    price.Discount
	LiveMode bool
}

// NewCheckout creates a new Checkout instance.
// ftcPrice - the current pricing policy and all available discounts.
// m - current membership used to deduce which offer is applicable for the next order.
func NewCheckout(ftcPrice price.FtcPrice, m reader.Membership) (Checkout, error) {

	// Find out which kind of order user is creating.
	orderKind, err := m.OrderKindOfOneTime(ftcPrice.Edition)
	if err != nil {
		return Checkout{}, err
	}

	// Find out applicable discount offer.
	discount := ftcPrice.ApplicableOffer(m.OfferKindsEnjoyed())

	return Checkout{
		Kind:     orderKind,
		Price:    ftcPrice.Price,
		Offer:    discount,
		LiveMode: true,
	}, nil
}

func (c Checkout) WithTest(t bool) Checkout {
	c.LiveMode = !t

	return c
}

func (c Checkout) Payable() price.Charge {
	charge := price.NewCharge(c.Price, c.Offer)
	if c.LiveMode {
		return charge
	}

	return charge.WithTest()
}
