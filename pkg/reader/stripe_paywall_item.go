package reader

import "github.com/FTChinese/subscription-api/pkg/price"

// StripePaywallItem combines stripe price and a list of
// applicable coupons.
type StripePaywallItem struct {
	Price   price.StripePrice    `json:"price"`
	Coupons []price.StripeCoupon `json:"coupons"`
}

func NewPaywallStripe(prices []price.StripePrice, coupons []price.StripeCoupon) []StripePaywallItem {
	groupedCoupons := groupCoupons(coupons)

	var result = make([]StripePaywallItem, 0)

	for _, sp := range prices {
		coupons, ok := groupedCoupons[sp.ID]
		// Introductory price should not contain coupons
		if !ok || sp.IsIntro() {
			coupons = []price.StripeCoupon{}
		}

		result = append(result, StripePaywallItem{
			Price:   sp,
			Coupons: coupons,
		})
	}

	return result
}

// Group a list of coupons by putting coupons with the same
// price id into separate list.
func groupCoupons(coupons []price.StripeCoupon) map[string][]price.StripeCoupon {
	var g = map[string][]price.StripeCoupon{}

	for _, coupon := range coupons {
		if coupon.PriceID.IsZero() {
			continue
		}

		found, ok := g[coupon.PriceID.String]
		if ok {
			found = append(found, coupon)
		} else {
			found = []price.StripeCoupon{
				coupon,
			}
		}

		g[coupon.PriceID.String] = found
	}

	return g
}

func (item StripePaywallItem) FindCoupon(id string) price.StripeCoupon {
	if item.Price.IsIntro() {
		return price.StripeCoupon{}
	}

	for _, v := range item.Coupons {
		if v.ID == id && v.IsValid() {
			return v
		}
	}

	return price.StripeCoupon{}
}
