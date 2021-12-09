package pw

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
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

func (w Paywall) findFtcPrice(id string) (PaywallPrice, error) {
	for _, prod := range w.Products {
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

func (w Paywall) FindCheckoutItem(priceID string, offerID null.String) (price.CheckoutItem, error) {
	ftcPrice, err := w.findFtcPrice(priceID)
	if err != nil {
		return price.CheckoutItem{}, err
	}

	if offerID.IsZero() {
		return price.CheckoutItem{
			Price: ftcPrice.Price,
			Offer: price.Discount{},
		}, nil
	}

	offer, _ := ftcPrice.Offers.FindValid(offerID.String)

	return price.CheckoutItem{
		Price: ftcPrice.Price,
		Offer: offer,
	}, nil
}

// StripePriceIDs collect all stripe price id present on paywall.
func (w Paywall) StripePriceIDs() []string {
	var ids = make([]string, 0)

	for _, pwProd := range w.Products {
		if pwProd.Introductory.StripePriceID.Valid {
			ids = append(ids, pwProd.Introductory.StripePriceID.String)
		}

		for _, pwPrice := range pwProd.Prices {
			if pwPrice.StripePriceID != "" {
				ids = append(ids, pwPrice.StripePriceID)
			}
		}
	}

	return ids
}
