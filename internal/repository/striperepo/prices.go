package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/stripe"
)

func (env Env) ListPrices(force bool) ([]stripe.Price, error) {
	if !force && PriceCache.Len() != 0 {
		return PriceCache.
			List(env.client.live), nil
	}

	sp, err := env.client.ListPrices()
	if err != nil {
		return nil, err
	}

	PriceCache.AddAll(sp)

	return PriceCache.
		List(env.client.live), nil
}

func (env Env) LoadPrice(id string, force bool) (stripe.Price, error) {
	if !force {
		p, ok := PriceCache.Find(id)
		if ok {
			return p, nil
		}
	}

	sp, err := env.client.RetrievePrice(id)
	if err != nil {
		return stripe.Price{}, err
	}

	return stripe.NewPrice(sp), nil
}

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
