package stripeclient

import (
	"context"
	"github.com/FTChinese/subscription-api/pkg/price"
	stripeSdk "github.com/stripe/stripe-go/v72"
	"golang.org/x/sync/semaphore"
	"runtime"
)

var (
	maxWorkers = runtime.GOMAXPROCS(0)
	sem        = semaphore.NewWeighted(int64(maxWorkers))
)

func (c Client) ListPrices() ([]*stripeSdk.Price, error) {
	iter := c.sc.Prices.List(&stripeSdk.PriceListParams{
		Active: stripeSdk.Bool(true),
		ListParams: stripeSdk.ListParams{
			Limit: stripeSdk.Int64(100),
		},
	})

	list := iter.PriceList()
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return list.Data, nil
}

// FetchPrice from stripe API.
// See https://stripe.com/docs/api/prices/retrieve.
// It seems the SDK handled it incorrectly since the API doc
// says no parameter is required.
func (c Client) FetchPrice(id string) (*stripeSdk.Price, error) {
	return c.sc.Prices.Get(id, nil)
}

func (c Client) FetchPricesOf(ids []string) (map[string]price.StripePrice, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()
	ctx := context.Background()

	var prices = make(map[string]price.StripePrice)
	var anyErr error

	for _, id := range ids {
		if err := sem.Acquire(ctx, 1); err != nil {
			sugar.Errorf("Failed to acquire semaphore: %v", err)
			break
		}

		go func(id string) {
			p, err := c.FetchPrice(id)
			if err != nil {
				sugar.Error(err)
				anyErr = err
			} else {
				prices[id] = price.NewStripePrice(p)
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

func (c Client) SetPriceMeta(id string, meta map[string]string) (*stripeSdk.Price, error) {
	return c.sc.Prices.Update(id, &stripeSdk.PriceParams{
		Params: stripeSdk.Params{
			Metadata: meta,
		},
	})
}
