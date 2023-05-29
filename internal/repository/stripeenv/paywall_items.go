package stripeenv

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// LoadOrFetchPaywallPrices loads all stripe prices for a list of ids.
func (env Env) LoadOrFetchPaywallPrices(refresh bool, live bool) ([]price.StripePrice, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	sugar.Infof("Load stripe prices from db")
	list, err := env.ListPaywallPrices(live)
	if err != nil {
		return nil, err
	}

	if !refresh {
		return list, nil
	}

	ids := make([]string, 0)
	for _, v := range list {
		ids = append(ids, v.ID)
	}

	sugar.Infof("Fetch from Stripe API...")
	priceList, err := env.Client.FetchPricesOf(ids)
	if err != nil {
		return nil, err
	}

	return priceList, nil
}

func (env Env) LoadOrFetchPaywallItems(refresh bool, live bool) ([]reader.StripePaywallItem, error) {
	pCh := make(chan pkg.AsyncResult[[]price.StripePrice])
	cCh := make(chan pkg.AsyncResult[[]price.StripeCoupon])

	go func() {
		defer close(pCh)

		prices, err := env.LoadOrFetchPaywallPrices(refresh, live)

		pCh <- pkg.AsyncResult[[]price.StripePrice]{
			Value: prices,
			Err:   err,
		}
	}()

	go func() {
		defer close(cCh)

		coupons, err := env.ListPaywallCoupons(live)

		cCh <- pkg.AsyncResult[[]price.StripeCoupon]{
			Value: coupons,
			Err:   err,
		}
	}()

	pRes, cRes := <-pCh, <-cCh

	if pRes.Err != nil {
		return nil, pRes.Err
	}

	if cRes.Err != nil {
		return nil, cRes.Err
	}

	return reader.NewPaywallStripe(pRes.Value, cRes.Value), nil
}

// LoadCheckoutItem from database, or from Stripe API if not found in database.
func (env Env) LoadCheckoutItem(params stripe.SubsParams, live bool) (reader.CartItemStripe, error) {
	recurring, err := env.LoadOrFetchPrice(params.PriceID, false, live)
	if err != nil {
		return reader.CartItemStripe{}, err
	}

	var introPrice price.StripePrice
	var coupon price.StripeCoupon
	if params.IntroductoryPriceID.Valid {
		introPrice, err = env.LoadOrFetchPrice(params.IntroductoryPriceID.String, false, live)
		if err != nil {
			return reader.CartItemStripe{}, err
		}
	}

	if params.CouponID.Valid {
		coupon, err = env.LoadOrFetchCoupon(params.CouponID.String, false, live)
		if err != nil {
			return reader.CartItemStripe{}, err
		}
	}

	return reader.CartItemStripe{
		Recurring:    recurring,
		Introductory: introPrice,
		Coupon:       coupon,
	}, nil
}
