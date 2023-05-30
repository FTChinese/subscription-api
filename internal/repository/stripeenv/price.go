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

	sp := price.NewStripePrice(rawPrice)

	err = env.UpsertPrice(sp)
	if err != nil {
		sugar.Error(err)
	}

	ok, err := env.IsPriceOnPaywall(id)
	if err != nil {
		sugar.Error(err)
	}

	sp.OnPaywall = ok

	return sp, nil
}
