package shared

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
)

// RetrievePaywallPrice retrieves a single row by plan id.
func (env PaywallCommon) RetrievePaywallPrice(id string, live bool) (pw.PaywallPrice, error) {
	var p pw.PaywallPrice
	err := env.dbs.Read.Get(
		&p,
		pw.StmtSelectPaywallPrice,
		id,
		live)

	if err != nil {
		return pw.PaywallPrice{}, err
	}

	return p, nil
}

type asyncPriceResult struct {
	value pw.PaywallPrice
	error error
}

func (env PaywallCommon) asyncLoadPrice(id string, live bool) <-chan asyncPriceResult {
	c := make(chan asyncPriceResult)

	go func() {
		defer close(c)

		p, err := env.RetrievePaywallPrice(id, live)

		c <- asyncPriceResult{
			value: p,
			error: err,
		}
	}()

	return c
}

// LoadDiscount retrieve a single row of discount
func (env PaywallCommon) LoadDiscount(id string) (price.Discount, error) {
	var d price.Discount
	err := env.dbs.Read.Get(
		&d,
		price.StmtSelectDiscount,
		id)
	if err != nil {
		return price.Discount{}, err
	}

	return d, nil
}

type asyncDiscountResult struct {
	value price.Discount
	error error
}

// load discount asynchronously. If id is empty, returns immediately.
func (env PaywallCommon) asyncLoadDiscount(id string) <-chan asyncDiscountResult {
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
func (env PaywallCommon) LoadCheckoutItem(params pw.CartParams, live bool) (price.CheckoutItem, error) {

	if params.DiscountID.IsZero() {
		pwp, err := env.RetrievePaywallPrice(params.PriceID, live)
		if err != nil {
			return price.CheckoutItem{}, err
		}

		return price.CheckoutItem{
			Price: pwp.Price,
			Offer: price.Discount{},
		}, nil
	}

	priceCh, discCh := env.asyncLoadPrice(params.PriceID, live), env.asyncLoadDiscount(params.DiscountID.String)

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
