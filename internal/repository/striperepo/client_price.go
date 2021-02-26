package striperepo

import "github.com/stripe/stripe-go/v72"

func (c Client) ListPrices() ([]*stripe.Price, error) {
	iter := c.sc.Prices.List(&stripe.PriceListParams{
		Active: stripe.Bool(true),
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
