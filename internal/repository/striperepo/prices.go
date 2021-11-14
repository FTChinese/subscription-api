package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/stripe"
)

func (env Env) ListPrices(force bool) ([]stripe.Price, error) {
	if !force && StripePriceCache.Len() != 0 {
		return StripePriceCache.
			List(env.client.live), nil
	}

	sp, err := env.client.ListPrices()
	if err != nil {
		return nil, err
	}

	StripePriceCache.AddAll(sp)

	return StripePriceCache.
		List(env.client.live), nil
}
