//go:build !production

package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
)

var MockPwPriceStdYear = PaywallPrice{
	FtcPrice: price.MockFtcStdYearPrice,
	Offers:   price.MockFtcStdYearOffers,
}

var MockPwPriceStdMonth = PaywallPrice{
	FtcPrice: price.MockFtcStdMonthPrice,
	Offers:   price.MockFtcStdMonthOffers,
}

var MockPwPricePrm = PaywallPrice{
	FtcPrice: price.MockFtcPrmPrice,
	Offers:   price.MockFtcPrmOffers,
}

var MockPaywall = Paywall{
	PaywallDoc: PaywallDoc{
		ID: 0,
		DailyBanner: BannerJSON{
			ID:           ids.BannerID(),
			Heading:      gofakeit.Word(),
			SubHeading:   null.String{},
			CoverURL:     null.String{},
			Content:      null.String{},
			Terms:        null.String{},
			ChronoPeriod: dt.ChronoPeriod{},
		},
		PromoBanner: BannerJSON{
			ID:           "",
			Heading:      "",
			SubHeading:   null.String{},
			CoverURL:     null.String{},
			Content:      null.String{},
			Terms:        null.String{},
			ChronoPeriod: dt.ChronoPeriod{},
		},
		LiveMode:   false,
		CreatedUTC: chrono.TimeNow(),
	},
	Products: []PaywallProduct{
		{
			Product: Product{
				ID:       ids.ProductID(),
				Active:   true,
				LiveMode: false,
				ProductParams: ProductParams{
					CreatedBy:   "anonymous",
					Description: null.String{},
					Heading:     "Standard Edition",
					SmallPrint:  null.String{},
					Tier:        enum.TierStandard,
				},
				Introductory: price.FtcPriceJSON{
					FtcPrice: price.MockFtcStdIntroPrice,
				},
				CreatedUTC: chrono.TimeNow(),
				UpdatedUTC: chrono.Time{},
			},
			Prices: []PaywallPrice{
				{
					FtcPrice: price.MockFtcStdYearPrice,
					Offers:   price.MockFtcStdYearOffers,
				},
				{
					FtcPrice: price.MockFtcStdMonthPrice,
					Offers:   price.MockFtcStdMonthOffers,
				},
			},
		},
		{
			Product: Product{
				ID:       ids.ProductID(),
				Active:   true,
				LiveMode: false,
				ProductParams: ProductParams{
					CreatedBy:   "",
					Description: null.String{},
					Heading:     "Premium Edition",
					SmallPrint:  null.String{},
					Tier:        enum.TierPremium,
				},
				Introductory: price.FtcPriceJSON{},
				CreatedUTC:   chrono.TimeNow(),
				UpdatedUTC:   chrono.Time{},
			},
			Prices: []PaywallPrice{
				{
					FtcPrice: price.MockFtcPrmPrice,
					Offers:   price.MockFtcPrmOffers,
				},
			},
		},
	},
	Stripe: []StripePaywallItem{
		{
			Price:   price.MockStripeStdIntroPrice,
			Coupons: nil,
		},
		{
			Price:   price.MockStripeStdYearPrice,
			Coupons: price.MockStripeStdYearCoupons,
		},
		{
			Price:   price.MockStripeStdMonthPrice,
			Coupons: nil,
		},
		{
			Price:   price.MockStripePrmPrice,
			Coupons: nil,
		},
	},
}
