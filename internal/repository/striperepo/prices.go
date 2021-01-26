package striperepo

import "github.com/FTChinese/subscription-api/pkg/stripe"

func (env Env) ListPrices(force bool) ([]stripe.Price, error) {
	if !force && stripe.PriceCache.Len() != 0 {
		return stripe.
			PriceCache.
			List(env.client.live), nil
	}

	sp, err := env.client.ListPrices()
	if err != nil {
		return nil, err
	}

	stripe.PriceCache.AddAll(sp)

	return stripe.
		PriceCache.
		List(env.client.live), nil
}

func (env Env) GetPrice(id string) (stripe.Price, error) {
	price, err := stripe.PriceCache.Find(id, env.client.live)
	if err == nil {
		return price, nil
	}

	sp, err := env.client.GetPrice(id)
	if err != nil {
		return stripe.Price{}, err
	}

	err = stripe.PriceCache.Upsert(sp)
	if err != nil {
		return stripe.Price{}, err
	}

	return stripe.PriceCache.Find(id, env.client.live)
}
