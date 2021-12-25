package pw

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/price"
)

type Paywall struct {
	PaywallDoc
	Products []PaywallProduct `json:"products"`
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
					Price:  prod.Introductory.Price,
					Offers: nil,
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

func (w Paywall) FindCheckoutItem(params CartParams) (price.CheckoutItem, error) {
	pwPrice, err := w.findPrice(params.PriceID)
	if err != nil {
		return price.CheckoutItem{}, err
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
