package reader

import (
	"database/sql"

	"github.com/FTChinese/subscription-api/pkg/price"
)

type Paywall struct {
	PaywallDoc
	Products  []PaywallProduct    `json:"products"`
	FTCPrices []PaywallPrice      `json:"ftcPrices"`
	Stripe    []StripePaywallItem `json:"stripe"`
}

func NewPaywall(pwb PaywallDoc, p []PaywallProduct) Paywall {
	return Paywall{
		PaywallDoc: pwb,
		Products:   p,
	}
}

// Normalize keeps backward compatibility.
// TODO: sync stripe ids from Stripe field to FTCprices;
// then sync FTCPrices to each products' prices field.
func (w Paywall) Normalize() Paywall {
	stripeIds := map[string]string{}

	for _, v := range w.Stripe {
		stripeIds[v.Price.Edition().String()] = v.Price.ID
	}

	for i, prod := range w.Products {
		for j, price := range prod.Prices {
			if prod.Introductory.ID != "" {
				sid := stripeIds[prod.Introductory.Edition.String()]

				w.Products[i].Introductory.StripePriceID = sid

				w.FTCPrices = append(w.FTCPrices, PaywallPrice{
					FtcPrice: w.Products[i].Introductory.FtcPrice,
				})
			}

			sid := stripeIds[price.Edition.String()]
			w.Products[i].Prices[j].StripePriceID = sid

			w.FTCPrices = append(w.FTCPrices, w.Products[i].Prices...)
		}
	}

	return w
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
