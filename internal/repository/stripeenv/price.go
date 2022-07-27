package stripeenv

import (
	"github.com/FTChinese/subscription-api/pkg/price"
)

// LoadOrFetchPrice tris to retrieve a price from
// db; then hit Stripe API if not found.
func (env Env) LoadOrFetchPrice(id string, refresh bool, live bool) (price.StripePrice, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	if !refresh {
		p, err := env.RetrievePrice(id, live)
		if err == nil {
			return p, err
		}

		sugar.Error(err)
	}

	rawPrice, err := env.Client.FetchPrice(id)
	if err != nil {
		return price.StripePrice{}, err
	}

	return price.NewStripePrice(rawPrice), nil
}

// ListPricesCompat fetches stripe prices from cache or from API.
// Deprecated.
func (env Env) ListPricesCompat(live bool, refresh bool) ([]price.StripePrice, error) {
	if !refresh && env.Cache.Len() != 0 {
		return env.Cache.
			List(live), nil
	}

	sp, err := env.Client.ListPrices()
	if err != nil {
		return nil, err
	}

	env.Cache.AddAll(sp)

	return env.Cache.
		List(live), nil
}
