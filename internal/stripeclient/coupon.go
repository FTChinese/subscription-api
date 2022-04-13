package stripeclient

import stripeSdk "github.com/stripe/stripe-go/v72"

func (c Client) FetchCoupon(id string) (*stripeSdk.Coupon, error) {
	return c.sc.Coupons.Get(id, nil)
}

func (c Client) UpdateCoupon(id string, m map[string]string) (*stripeSdk.Coupon, error) {
	p := &stripeSdk.CouponParams{
		Params: stripeSdk.Params{
			Metadata: m,
		},
	}

	return c.sc.Coupons.Update(id, p)
}
