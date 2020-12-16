package striperepo

import "github.com/FTChinese/subscription-api/pkg/stripe"

func (env Env) ListPrices() []stripe.Price {
	if stripe.PriceStore.Len() != 0 {
		return stripe.
			PriceStore.
			List(env.client.live)
	}

	sp := env.client.ListPrices()

	stripe.PriceStore.AddAll(sp)

	return stripe.
		PriceStore.
		List(env.client.live)
}

func (env Env) RefreshPrices() []stripe.Price {
	sp := env.client.ListPrices()

	stripe.PriceStore.AddAll(sp)

	return stripe.
		PriceStore.
		List(env.client.live)
}
