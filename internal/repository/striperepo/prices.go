package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/stripe"
)

func (env Env) LoadCheckoutItem(params stripe.SubsParams) (stripe.CheckoutItem, error) {
	p, err := env.LoadPrice(params.PriceID, false)
	if err != nil {
		return stripe.CheckoutItem{}, err
	}

	if params.IntroductoryPriceID.IsZero() {
		return stripe.CheckoutItem{
			Price:        p,
			Introductory: stripe.Price{},
		}, nil
	}

	introPrice, err := env.LoadPrice(params.IntroductoryPriceID.String, false)
	if err != nil {
		return stripe.CheckoutItem{}, err
	}

	return stripe.CheckoutItem{
		Price:        p,
		Introductory: introPrice,
	}, nil
}
