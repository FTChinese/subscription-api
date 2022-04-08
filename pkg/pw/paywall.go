package pw

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/price"
)

type Paywall struct {
	PaywallDoc
	Products     []PaywallProduct             `json:"products"`
	StripePrices map[string]price.StripePrice `json:"stripePrices"`
}

func NewPaywall(pwb PaywallDoc, p []PaywallProduct) Paywall {
	return Paywall{
		PaywallDoc: pwb,
		Products:   p,
	}
}

func (w Paywall) findPrice(id string) (PaywallPrice, error) {

	for _, prod := range w.Products {
		if !prod.Introductory.IsZero() {
			if prod.Introductory.ID == id {
				return PaywallPrice{
					FtcPrice: prod.Introductory.FtcPrice,
					Offers:   nil,
				}, nil
			}
		}

		for _, v := range prod.Prices {
			if v.ID == id {
				return v, nil
			}
		}
	}

	return PaywallPrice{}, errors.New("the requested price is not found")
}

// FindPriceByEdition tries to find a price for a specific edition.
// Deprecated.
func (w Paywall) FindPriceByEdition(e price.Edition) (PaywallPrice, error) {
	for _, prod := range w.Products {
		for _, p := range prod.Prices {
			if p.Edition == e {
				return p, nil
			}
		}
	}

	return PaywallPrice{}, sql.ErrNoRows
}

// BuildFtcCartItem constructs an FTC price checkout item
// from cached data.
func (w Paywall) BuildFtcCartItem(params FtcCartParams) (CartItemFtc, error) {
	pwPrice, err := w.findPrice(params.PriceID)
	if err != nil {
		return CartItemFtc{}, err
	}

	return pwPrice.CheckoutItem(params.DiscountID)
}

// StripePriceIDs collect all stripe price id present on paywall.
func (w Paywall) StripePriceIDs() []string {
	var ids = make([]string, 0)

	for _, pwProd := range w.Products {
		if pwProd.Introductory.StripePriceID != "" {
			ids = append(ids, pwProd.Introductory.StripePriceID)
		}

		for _, pwPrice := range pwProd.Prices {
			if pwPrice.StripePriceID != "" {
				ids = append(ids, pwPrice.StripePriceID)
			}
		}
	}

	return ids
}

// BuildStripeCartItem constructs a Stripe checkout item
// from cached data.
func (w Paywall) BuildStripeCartItem(params StripeSubsParams) (CartItemStripe, error) {
	recurring, ok := w.StripePrices[params.PriceID]
	if !ok {
		return CartItemStripe{}, fmt.Errorf("stripe price %s not found", params.PriceID)
	}

	var intro price.StripePrice
	if params.IntroductoryPriceID.Valid {
		intro, ok = w.StripePrices[params.IntroductoryPriceID.String]
		if !ok {
			return CartItemStripe{}, fmt.Errorf("stripe price %s not found", params.IntroductoryPriceID.String)
		}
	}

	return CartItemStripe{
		Recurring:    recurring,
		Introductory: intro,
	}, nil
}
