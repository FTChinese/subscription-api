package products

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

type asyncPriceResult struct {
	value price.FtcPrice
	error error
}

func (env Env) asyncLoadPrice(id string) <-chan asyncPriceResult {
	c := make(chan asyncPriceResult)

	go func() {
		defer close(c)

		p, err := env.RetrieveFtcPrice(id)

		c <- asyncPriceResult{
			value: p,
			error: err,
		}
	}()

	return c
}

type asyncDiscountResult struct {
	value price.Discount
	error error
}

// load discount asynchronously. If id is empty, returns immediately.
func (env Env) asyncLoadDiscount(id string) <-chan asyncDiscountResult {
	c := make(chan asyncDiscountResult)

	if id == "" {
		c <- asyncDiscountResult{
			value: price.Discount{},
			error: nil,
		}
	} else {
		go func() {
			defer close(c)

			d, err := env.LoadDiscount(id)

			c <- asyncDiscountResult{
				value: d,
				error: err,
			}
		}()
	}

	return c
}

// LoadCheckoutItem loads a price and a discount from db.
func (env Env) LoadCheckoutItem(priceID string, discountID null.String) (price.CheckoutItem, error) {
	priceCh, discCh := env.asyncLoadPrice(priceID), env.asyncLoadDiscount(discountID.String)

	priceResult, discResult := <-priceCh, <-discCh
	if priceResult.error != nil {
		return price.CheckoutItem{}, priceResult.error
	}

	if discResult.error != nil {
		return price.CheckoutItem{}, discResult.error
	}

	return price.CheckoutItem{
		Price: priceResult.value.Price,
		Offer: discResult.value,
	}, nil
}
