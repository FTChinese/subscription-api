package reader

import (
	"database/sql"
	"errors"

	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

type Paywall struct {
	PaywallDoc
	Products  []PaywallProduct    `json:"products"`
	FTCPrices []PaywallPrice      `json:"ftcPrices"`
	Stripe    []StripePaywallItem `json:"stripe"`
}

// Deprecated.
func NewPaywall(pwb PaywallDoc, p []PaywallProduct) Paywall {
	return Paywall{
		PaywallDoc: pwb,
		Products:   p,
	}
}

func BuildFtcPaywall(pwd PaywallDoc, products []Product, prices []PaywallPrice) Paywall {
	return Paywall{
		PaywallDoc: pwd,
		Products:   NewPaywallProductsV2(products, prices),
		FTCPrices:  prices,
	}
}

// Normalize keeps backward compatibility.
// Sync stripe ids from Stripe field to each products' prices field.
func (w Paywall) Normalize() Paywall {
	stripeIds := map[string]string{}

	// Collection stripe ids into a map.
	for _, v := range w.Stripe {
		stripeIds[v.Price.Edition().String()] = v.Price.ID
	}

	// Loop over each product.
	for i, prod := range w.Products {
		// Link introductory price
		if prod.Introductory.ID != "" {
			sid := stripeIds[prod.Introductory.Edition.String()]

			w.Products[i].Introductory.StripePriceID = sid
		}

		// Loop over each price.
		for j, price := range prod.Prices {

			sid := stripeIds[price.Edition.String()]

			w.Products[i].Prices[j].StripePriceID = sid
		}
	}

	return w
}

func (w Paywall) FindFtcPrice(priceID string) (PaywallPrice, error) {
	for _, v := range w.FTCPrices {
		if v.ID == priceID {
			return v, nil
		}
	}

	return PaywallPrice{}, errors.New("the requested price is not found")
}

// CartItemFtc finds a price and an optional discount.
func (w Paywall) CartItemFtc(priceID string, discountID null.String) (CartItemFtc, error) {
	pwPrice, err := w.FindFtcPrice(priceID)

	if err != nil {
		return CartItemFtc{}, err
	}

	return pwPrice.BuildCartItem(discountID)
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
