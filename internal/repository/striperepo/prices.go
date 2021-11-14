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
