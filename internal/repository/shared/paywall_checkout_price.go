package shared

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

// RetrieveFtcPrice retrieves a single row by plan id.
func (env PaywallCommon) RetrieveFtcPrice(id string, live bool) (price.FtcPrice, error) {
	var p price.FtcPrice
	err := env.DBs.Read.Get(
		&p,
		price.StmtFtcPrice,
		id,
		live)

	if err != nil {
		return price.FtcPrice{}, err
	}

	return p, nil
}

type asyncPriceResult struct {
	value price.FtcPrice
	error error
}

func (env PaywallCommon) asyncLoadPrice(id string, live bool) <-chan asyncPriceResult {
	c := make(chan asyncPriceResult)

	go func() {
		defer close(c)

		p, err := env.RetrieveFtcPrice(id, live)

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
	err := env.DBs.Read.Get(
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
func (env PaywallCommon) LoadCheckoutItem(priceID string, discountID null.String, live bool) (price.CheckoutItem, error) {
	priceCh, discCh := env.asyncLoadPrice(priceID, live), env.asyncLoadDiscount(discountID.String)

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
