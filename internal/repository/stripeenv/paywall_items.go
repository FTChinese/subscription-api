package stripeenv

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// loadOrFetchPaywallPrices loads all stripe prices for a list of ids.
func (env Env) loadOrFetchPaywallPrices(ids []string, refresh bool) ([]price.StripePrice, error) {
	var diff []string
	if !refresh {
		list, err := env.RetrievePaywallPrices(ids)
		if err != nil {
			diff = ids
		} else {
			diff = isAllPriceRetrieved(ids, list)
		}

		if len(diff) == 0 {
			return list, nil
		}
	}

	pricesMap, err := env.Client.FetchPricesOf(ids)
	if err != nil {
		return nil, err
	}

	priceList := make([]price.StripePrice, 0)
	for _, v := range pricesMap {
		priceList = append(priceList, v)
	}

	return priceList, nil
}

func isAllPriceRetrieved(ids []string, prices []price.StripePrice) []string {
	var result []string
	var received map[string]interface{}
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

func (env Env) LoadOrFetchPaywallItems(priceIDs []string, refresh bool) ([]reader.StripePaywallItem, error) {
	prices, err := env.loadOrFetchPaywallPrices(priceIDs, refresh)
	if err != nil {
		return nil, err
	}

	coupons, err := env.RetrievePaywallCoupons(priceIDs)
	if err != nil {
		return nil, err
	}

	return reader.NewPaywallStripe(prices, coupons), nil
}

// LoadCheckoutItem from database, or from Stripe API if not found in database.
func (env Env) LoadCheckoutItem(params stripe.SubsParams) (reader.CartItemStripe, error) {
	recurring, err := env.LoadOrFetchPrice(params.PriceID, false)
	if err != nil {
		return reader.CartItemStripe{}, err
	}

	var introPrice price.StripePrice
	var coupon price.StripeCoupon
	if params.IntroductoryPriceID.Valid {
		introPrice, err = env.LoadOrFetchPrice(params.IntroductoryPriceID.String, false)
		if err != nil {
			return reader.CartItemStripe{}, err
		}
	}

	if params.CouponID.Valid {
		coupon, err = env.LoadOrFetchCoupon(params.CouponID.String, false)
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
