package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/price"
)

func (env Env) ListPrices(force bool) ([]price.StripePrice, error) {
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
