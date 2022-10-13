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

func (c Client) fetchPrice(id string, s *semaphore.Weighted, priceCh chan<- price.StripePrice) {
	defer s.Release(1)

	p, err := c.FetchPrice(id)
	// TODO: properly handle error
	if err != nil {
		return
	}

	priceCh <- price.NewStripePrice(p)
}

func (c Client) FetchPricesOf(ids []string) ([]price.StripePrice, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()
	ctx := context.Background()

	priceCh := make(chan price.StripePrice)

	for _, id := range ids {
		if err := sem.Acquire(ctx, 1); err != nil {
			sugar.Errorf("Failed to acquire semaphore: %v", err)
			break
		}

		sugar.Infof("Start a new goroutine to fetch stripe price %s", id)
		go c.fetchPrice(id, sem, priceCh)
	}

	go func() {
		if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
			sugar.Error(err)
		} else {
			close(priceCh)
		}
	}()

	var prices []price.StripePrice
	for p := range priceCh {
		prices = append(prices, p)
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
