package pw

import (
	"errors"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

type Paywall struct {
	PaywallDoc
	Products []Product `json:"products"`
}

func NewPaywall(pwb PaywallDoc, p []Product) Paywall {
	return Paywall{
		PaywallDoc: pwb,
		Products:   p,
	}
}

func (w Paywall) findFtcPrice(id string) (price.FtcPrice, error) {
	for _, prod := range w.Products {
		for _, v := range prod.Prices {
			if v.ID == id {
				return v, nil
			}
		}
	}

	return price.FtcPrice{}, errors.New("the requested price is not found")
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
