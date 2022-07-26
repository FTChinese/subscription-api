package stripeenv

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// loadOrFetchPaywallPrices loads all stripe prices for a list of ids.
func (env Env) loadOrFetchPaywallPrices(ids []string, refresh bool, live bool) ([]price.StripePrice, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	var diff []string
	if !refresh {
		sugar.Infof("Load stripe prices from db")
		list, err := env.RetrievePaywallPrices(ids, live)
		if err != nil {
			sugar.Error(err)
			diff = ids
		} else {
			diff = isAllPriceRetrieved(ids, list)
		}

		if len(diff) == 0 {
			return list, nil
		}

		sugar.Infof("Some of stripes not found in db: %s", ids)
	}

	sugar.Infof("Fetch from Stripe API...")
	priceList, err := env.Client.FetchPricesOf(ids)
	if err != nil {
		return nil, err
	}

	return priceList, nil
}

func isAllPriceRetrieved(ids []string, prices []price.StripePrice) []string {
	var result = make([]string, 0)
	var received = make(map[string]interface{})
	for _, v := range prices {
		received[v.ID] = nil
	}

	for _, id := range ids {
		_, ok := received[id]
		if !ok {
			result = append(result, id)
		}
	}

	return result
}

func (env Env) LoadOrFetchPaywallItems(priceIDs []string, refresh bool, live bool) ([]reader.StripePaywallItem, error) {
	prices, err := env.loadOrFetchPaywallPrices(priceIDs, refresh, live)
	if err != nil {
		return nil, err
	}

	coupons, err := env.RetrievePaywallCoupons(priceIDs, live)
	if err != nil {
		return nil, err
	}

	return reader.NewPaywallStripe(prices, coupons), nil
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
