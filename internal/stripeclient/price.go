package stripeclient

import (
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

// FetchPrice from stripe API.
// See https://stripe.com/docs/api/prices/retrieve.
// It seems the SDK handled it incorrectly since the API doc
// says no parameter is required.
func (c Client) FetchPrice(id string) (*stripeSdk.Price, error) {
	return c.sc.Prices.Get(id, nil)
}

func (c Client) SetPriceMeta(id string, meta map[string]string) (*stripeSdk.Price, error) {
	return c.sc.Prices.Update(id, &stripeSdk.PriceParams{
		Params: stripeSdk.Params{
			Metadata: meta,
		},
	})
}
