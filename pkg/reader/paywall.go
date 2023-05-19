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
