package striperepo

import "github.com/FTChinese/subscription-api/pkg/stripe"

func (env Env) ListPrices() ([]stripe.Price, error) {
	if stripe.PriceStore.Len() != 0 {
		return stripe.
			PriceStore.
			List(env.client.live), nil
	}

	sp, err := env.client.ListPrices()
	if err != nil {
		return nil, err
	}

	stripe.PriceStore.AddAll(sp)

	return stripe.
		PriceStore.
		List(env.client.live), nil
}

func (env Env) RefreshPrices() ([]stripe.Price, error) {
	sp, err := env.client.ListPrices()
	if err != nil {
		return nil, err
	}

	stripe.PriceStore.AddAll(sp)

	return stripe.
		PriceStore.
		List(env.client.live), nil
}
