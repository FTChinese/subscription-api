package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/stripe/stripe-go/v72"
)

func (c Client) ListPrices() ([]*stripe.Price, error) {
	iter := c.sc.Prices.List(&stripe.PriceListParams{
		Active: stripe.Bool(true),
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(100),
		},
	})

	list := iter.PriceList()
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return list.Data, nil
}

func (c Client) GetPrice(id string) (*stripe.Price, error) {
	return c.sc.Prices.Get(id, &stripe.PriceParams{
		Active: stripe.Bool(true),
	})
}

func (c Client) GetAndCachePrice(id string) (price.StripePrice, error) {
	sp, err := c.GetPrice(id)
	if err != nil {
		return price.StripePrice{}, err
	}

	p := price.NewStripePrice(sp)
	price.StripePriceCache.Upsert(p)

	return p, nil
}
