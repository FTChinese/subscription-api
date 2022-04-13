package reader

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"reflect"
	"testing"
)

var couponAStdYear = price.StripeCoupon{
	ID: "coupon-a-standard-year",
	StripeCouponMeta: price.StripeCouponMeta{
		PriceID: null.StringFrom(price.MockStripeStdYearPrice.ID),
	},
}

var couponBStdYear = price.StripeCoupon{
	ID: "coupon-b-standard-year",
	StripeCouponMeta: price.StripeCouponMeta{
		PriceID: null.StringFrom(price.MockStripeStdYearPrice.ID),
	},
}

var couponAStdMonth = price.StripeCoupon{
	ID: "coupon-a-standard-month",
	StripeCouponMeta: price.StripeCouponMeta{
		PriceID: null.StringFrom(price.MockStripeStdMonthPrice.ID),
	},
}

var couponBStdMonth = price.StripeCoupon{
	ID: "coupon-b-standard-month",
	StripeCouponMeta: price.StripeCouponMeta{
		PriceID: null.StringFrom(price.MockStripeStdMonthPrice.ID),
	},
}

var couponAPrm = price.StripeCoupon{
	ID: "coupon-a-premium",
	StripeCouponMeta: price.StripeCouponMeta{
		PriceID: null.StringFrom(price.MockStripePrmPrice.ID),
	},
}

var couponBPrm = price.StripeCoupon{
	ID: "coupon-b-premium",
	StripeCouponMeta: price.StripeCouponMeta{
		PriceID: null.StringFrom(price.MockStripePrmPrice.ID),
	},
}

func TestNewPaywallStripe(t *testing.T) {
	type args struct {
		prices  []price.StripePrice
		coupons []price.StripeCoupon
	}
	tests := []struct {
		name string
		args args
		want []StripePaywallItem
	}{
		{
			name: "",
			args: args{
				prices: []price.StripePrice{
					price.MockStripeStdYearPrice,
					price.MockStripeStdMonthPrice,
					price.MockStripePrmPrice,
				},
				coupons: []price.StripeCoupon{
					couponAStdYear,
					couponBStdYear,
					couponAStdMonth,
					couponBStdMonth,
					couponAPrm,
					couponBPrm,
				},
			},
			want: []StripePaywallItem{
				{
					Price: price.MockStripeStdYearPrice,
					Coupons: []price.StripeCoupon{
						couponAStdYear,
						couponBStdYear,
					},
				},
				{
					Price: price.MockStripeStdMonthPrice,
					Coupons: []price.StripeCoupon{
						couponAStdMonth,
						couponBStdMonth,
					},
				},
				{
					Price: price.MockStripePrmPrice,
					Coupons: []price.StripeCoupon{
						couponAPrm,
						couponBPrm,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPaywallStripe(tt.args.prices, tt.args.coupons); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPaywallStripe() = %v, want %v", got, tt.want)
			}
		})
	}
}
