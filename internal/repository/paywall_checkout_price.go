package repository

import (
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// RetrievePaywallPrice retrieves a single row by plan id.
func (repo PaywallRepo) RetrievePaywallPrice(id string, live bool) (reader.PaywallPrice, error) {
	var p reader.PaywallPrice
	err := repo.dbs.Read.Get(
		&p,
		reader.StmtSelectPaywallPrice,
		id,
		live)

	if err != nil {
		return reader.PaywallPrice{}, err
	}

	return p, nil
}

type asyncPriceResult struct {
	value reader.PaywallPrice
	error error
}

func (repo PaywallRepo) asyncLoadPrice(id string, live bool) <-chan asyncPriceResult {
	c := make(chan asyncPriceResult)

	go func() {
		defer close(c)

		p, err := repo.RetrievePaywallPrice(id, live)

		c <- asyncPriceResult{
			value: p,
			error: err,
		}
	}()

	return c
}

// LoadDiscount retrieve a single row of discount
func (repo PaywallRepo) LoadDiscount(id string) (price.Discount, error) {
	var d price.Discount
	err := repo.dbs.Read.Get(
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
func (repo PaywallRepo) asyncLoadDiscount(id string) <-chan asyncDiscountResult {
	c := make(chan asyncDiscountResult)

	if id == "" {
		c <- asyncDiscountResult{
			value: price.Discount{},
			error: nil,
		}
	} else {
		go func() {
			defer close(c)

			d, err := repo.LoadDiscount(id)

			c <- asyncDiscountResult{
				value: d,
				error: err,
			}
		}()
	}

	return c
}

// LoadCheckoutItem loads a price and a discount from db.
func (repo PaywallRepo) LoadCheckoutItem(params ftcpay.FtcCartParams, live bool) (reader.CartItemFtc, error) {

	if params.DiscountID.IsZero() {
		pwp, err := repo.RetrievePaywallPrice(params.PriceID, live)
		if err != nil {
			return reader.CartItemFtc{}, err
		}

		return reader.CartItemFtc{
			Price: pwp.FtcPrice,
			Offer: price.Discount{},
		}, nil
	}

	priceCh, discCh := repo.asyncLoadPrice(params.PriceID, live), repo.asyncLoadDiscount(params.DiscountID.String)

	priceResult, discResult := <-priceCh, <-discCh
	if priceResult.error != nil {
		return reader.CartItemFtc{}, priceResult.error
	}

	if discResult.error != nil {
		return reader.CartItemFtc{}, discResult.error
	}

	return reader.CartItemFtc{
		Price: priceResult.value.FtcPrice,
		Offer: discResult.value,
	}, nil
}
