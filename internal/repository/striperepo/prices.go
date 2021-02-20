package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/price"
)

func (env Env) ListPrices(force bool) ([]price.Price, error) {
	if !force && price.StripePriceCache.Len() != 0 {
		return price.StripePriceCache.
			List(env.client.live), nil
	}

	sp, err := env.client.ListPrices()
	if err != nil {
		return nil, err
	}

	price.StripePriceCache.AddAll(sp)

	return price.StripePriceCache.
		List(env.client.live), nil
}

func (env Env) GetPrice(id string) (price.Price, error) {
	cachedPrice, err := price.StripePriceCache.Find(id, env.client.live)
	if err == nil {
		return cachedPrice, nil
	}

	sp, err := env.client.GetPrice(id)
	if err != nil {
		return price.Price{}, err
	}

	err = price.StripePriceCache.Upsert(sp)
	if err != nil {
		return price.Price{}, err
	}

	return price.StripePriceCache.Find(id, env.client.live)
}
