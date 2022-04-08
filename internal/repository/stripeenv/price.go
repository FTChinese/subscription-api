package stripeenv

import (
	"context"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"golang.org/x/sync/semaphore"
	"runtime"
)

var (
	maxWorkers = runtime.GOMAXPROCS(0)
	sem        = semaphore.NewWeighted(int64(maxWorkers))
)

// LoadOrFetchPrice tris to retrieve a price from
// db; then hit Stripe API if not found.
func (env Env) LoadOrFetchPrice(id string, refresh bool) (price.StripePrice, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	if !refresh {
		p, err := env.RetrievePrice(id)
		if err == nil {
			return p, err
		}

		sugar.Error(err)
	}

	rawPrice, err := env.Client.FetchPrice(id)
	if err != nil {
		return price.StripePrice{}, err
	}

	return price.NewPrice(rawPrice), nil
}

func (env Env) LoadCachedPrice(id string, refresh bool) (price.StripePrice, error) {
	defer env.Logger.Sugar()
	sugar := env.Logger.Sugar()

	if !refresh {
		sugar.Infof("Loading stripe price %s from cache", id)
		p, ok := env.Cache.Find(id)
		if ok {
			sugar.Infof("Cached stripe price %s found", id)
			return p, nil
		}
	}

	p, err := env.LoadOrFetchPrice(id, refresh)
	if err != nil {
		sugar.Error(err)
		return price.StripePrice{}, err
	}

	env.Cache.Upsert(p)

	if p.IsFromStripe {
		go func() {
			err := env.UpsertPrice(p)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	return p, nil
}

// ListPrices loads all prices for a list of ids.
func (env Env) ListPrices(ids []string) (map[string]price.StripePrice, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()
	ctx := context.Background()

	var prices = make(map[string]price.StripePrice)
	var anyErr error

	for _, id := range ids {
		if err := sem.Acquire(ctx, 1); err != nil {
			sugar.Errorf("Failed to acquire semaphore: %v", err)
			break
		}

		go func(id string) {
			p, err := env.LoadOrFetchPrice(id, false)
			if err != nil {
				sugar.Error(err)
				anyErr = err
			} else {
				prices[id] = p
			}
			sem.Release(1)
		}(id)
	}

	if anyErr != nil {
		return nil, anyErr
	}

	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		sugar.Error(err)
		return nil, err
	}

	return prices, nil
}

// ListPricesCompat fetches stripe prices from cache or
// for API.
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

// LoadCheckoutItem from database, or from Stripe API if not found in database.
func (env Env) LoadCheckoutItem(params pw.StripeSubsParams) (pw.CartItemStripe, error) {
	recurring, err := env.LoadOrFetchPrice(params.PriceID, false)
	if err != nil {
		return pw.CartItemStripe{}, err
	}

	var introPrice price.StripePrice
	if params.IntroductoryPriceID.Valid {
		introPrice, err = env.LoadOrFetchPrice(params.IntroductoryPriceID.String, false)
		if err != nil {
			return pw.CartItemStripe{}, err
		}
	}

	return pw.CartItemStripe{
		Recurring:    recurring,
		Introductory: introPrice,
	}, nil
}
