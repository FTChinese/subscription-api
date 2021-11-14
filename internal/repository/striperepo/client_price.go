package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/stripe"
	stripeSdk "github.com/stripe/stripe-go/v72"
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

func (c Client) GetPrice(id string) (*stripeSdk.Price, error) {
	return c.sc.Prices.Get(id, &stripeSdk.PriceParams{
		Active: stripeSdk.Bool(true),
	})
}

func (c Client) GetAndCachePrice(id string) (stripe.Price, error) {
	sp, err := c.GetPrice(id)
	if err != nil {
		return stripe.Price{}, err
	}

	p := stripe.NewPrice(sp)
	PriceCache.Upsert(p)

	return p, nil
}
