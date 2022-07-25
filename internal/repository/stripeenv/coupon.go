package stripeenv

import "github.com/FTChinese/subscription-api/pkg/price"

func (env Env) LoadOrFetchCoupon(id string, refresh bool) (price.StripeCoupon, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	if !refresh {
		c, err := env.RetrieveCoupon(id)
		if err == nil {
			return c, nil
		}

		sugar.Error(err)
	}

	rawCoupon, err := env.Client.FetchCoupon(id)
	if err != nil {
		return price.StripeCoupon{}, err
	}

	return price.NewStripeCoupon(rawCoupon), nil
}

// ModifyCoupon modifies stripe coupon's metadata and then insert/update
// to FTC's database.
func (env Env) ModifyCoupon(id string, params price.StripeCouponMeta) (price.StripeCoupon, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	rawCoupon, err := env.Client.UpdateCoupon(id, params.ToMap())
	if err != nil {
		return price.StripeCoupon{}, err
	}

	coupon := price.NewStripeCoupon(rawCoupon)

	err = env.UpsertCoupon(coupon)
	if err != nil {
		sugar.Error(err)
	}

	return coupon, nil
}
