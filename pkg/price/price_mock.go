//go:build !production

package price

import (
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
)

var MockFtcStdIntroPrice = FtcPrice{
	ID: "price_x6rCUXxPC8tB",
	Edition: Edition{
		Tier:  enum.TierStandard,
		Cycle: enum.CycleNull,
	},
	Active:   true,
	Archived: false,
	Currency: "cny",
	Kind:     KindOneTime,
	LiveMode: false,
	Nickname: null.String{},
	PeriodCount: ColumnYearMonthDay{
		YearMonthDay: dt.YearMonthDay{
			Years:  0,
			Months: 1,
			Days:   0,
		},
	},
	ProductID: "prod_9xrJdHFq0wmq",
	// StripePriceID: "price_1Juuu2BzTK0hABgJTXiK4NTt",
	Title:      null.StringFrom("新会员首次订阅1元/月"),
	UnitAmount: 1,
	StartUTC:   chrono.TimeNow(),
	EndUTC:     chrono.TimeFrom(time.Now().AddDate(0, 0, 7)),
	CreatedUTC: chrono.TimeNow(),
}

var MockFtcStdYearPrice = FtcPrice{
	ID:       "price_WHc5ssjh6pqw",
	Edition:  MockEditionStdYear,
	Active:   true,
	Archived: false,
	Currency: CurrencyCNY,
	Kind:     KindRecurring,
	LiveMode: false,
	Nickname: null.String{},
	PeriodCount: ColumnYearMonthDay{
		YearMonthDay: dt.YearMonthDay{
			Years:  1,
			Months: 0,
			Days:   0,
		},
	},
	ProductID: "prod_9xrJdHFq0wmq",
	// StripePriceID: "price_1IM2nFBzTK0hABgJiIDeDIox",
	Title:      null.StringFrom("Standard Annual Edition"),
	UnitAmount: 298,
	StartUTC:   chrono.Time{},
	EndUTC:     chrono.Time{},
	CreatedUTC: chrono.TimeNow(),
}

var MockFtcStdYearOffers = []Discount{
	{
		ID:       "dsc_iirQArMFjBfs",
		LiveMode: false,
		Status:   DiscountStatusActive,
		DiscountParams: DiscountParams{
			Description: null.StringFrom("到期前续订享75折"),
			Kind:        OfferKindRetention,
			OverridePeriod: ColumnYearMonthDay{
				YearMonthDay: dt.YearMonthDay{
					Years:  0,
					Months: 0,
					Days:   0,
				},
			},
			Percent:   null.Int{},
			PriceOff:  null.FloatFrom(80),
			PriceID:   "price_WHc5ssjh6pqw",
			Recurring: true,
			TimeSlot:  dt.TimeSlot{},
			CreatedBy: "",
		},
		CreatedUTC: chrono.Time{},
	},
	{
		ID:       "dsc_Vn7686x357KY",
		LiveMode: false,
		Status:   DiscountStatusActive,
		DiscountParams: DiscountParams{
			Description: null.StringFrom("再次购买享八五折优惠"),
			Kind:        OfferKindWinBack,
			OverridePeriod: ColumnYearMonthDay{
				YearMonthDay: dt.YearMonthDay{
					Years:  0,
					Months: 0,
					Days:   0,
				},
			},
			Percent:   null.Int{},
			PriceOff:  null.FloatFrom(40),
			PriceID:   "price_WHc5ssjh6pqw",
			Recurring: true,
			TimeSlot:  dt.TimeSlot{},
			CreatedBy: "",
		},
		CreatedUTC: chrono.Time{},
	},
}

var MockFtcStdMonthPrice = FtcPrice{
	ID:       "price_v5E2WSqJymxe",
	Edition:  MockEditionStdMonth,
	Active:   true,
	Archived: false,
	Currency: CurrencyCNY,
	Kind:     KindRecurring,
	LiveMode: false,
	Nickname: null.String{},
	PeriodCount: ColumnYearMonthDay{
		YearMonthDay: dt.YearMonthDay{
			Years:  0,
			Months: 1,
			Days:   0,
		},
	},
	ProductID: "prod_9xrJdHFq0wmq",
	// StripePriceID: "price_1IM2mgBzTK0hABgJVH8o9Sjm",
	Title:      null.StringFrom("Standard Monthly Edition"),
	UnitAmount: 28,
	StartUTC:   chrono.Time{},
	EndUTC:     chrono.Time{},
	CreatedUTC: chrono.TimeNow(),
}

var MockFtcStdMonthOffers []Discount

var MockFtcPrmPrice = FtcPrice{
	ID:       "price_zsTj2TQ1h3jB",
	Edition:  MockEditionPrm,
	Active:   true,
	Archived: false,
	Currency: CurrencyCNY,
	Kind:     KindRecurring,
	LiveMode: false,
	Nickname: null.String{},
	PeriodCount: ColumnYearMonthDay{
		YearMonthDay: dt.YearMonthDay{
			Years:  1,
			Months: 0,
			Days:   0,
		},
	},
	ProductID: "prod_zSgOTS6DWLmu",
	// StripePriceID: "plan_FOde0uAr0V4WmT",
	Title:      null.StringFrom("Premium Annual Edition"),
	UnitAmount: 1998,
	StartUTC:   chrono.Time{},
	EndUTC:     chrono.Time{},
	CreatedUTC: chrono.TimeNow(),
}

var MockFtcPrmOffers = []Discount{
	{
		ID:       "dsc_m7f0nlLHdOoB",
		LiveMode: false,
		Status:   DiscountStatusActive,
		DiscountParams: DiscountParams{
			Description: null.StringFrom("现在续订享75折优惠"),
			Kind:        OfferKindRetention,
			OverridePeriod: ColumnYearMonthDay{
				YearMonthDay: dt.YearMonthDay{
					Years:  0,
					Months: 0,
					Days:   0,
				},
			},
			Percent:   null.Int{},
			PriceOff:  null.FloatFrom(500),
			PriceID:   "price_zsTj2TQ1h3jB",
			Recurring: true,
			TimeSlot:  dt.TimeSlot{},
			CreatedBy: "",
		},
		CreatedUTC: chrono.Time{},
	},
}

var MockStripeStdIntroPrice = StripePrice{
	IsFromStripe: false,
	ID:           "price_1Juuu2BzTK0hABgJTXiK4NTt",
	Active:       true,
	Currency:     "gbp",
	Kind:         KindOneTime,
	LiveMode:     false,
	Nickname:     "Trial period",
	ProductID:    "",
	PeriodCount: ColumnYearMonthDay{
		YearMonthDay: dt.YearMonthDay{
			Years:  0,
			Months: 1,
			Days:   0,
		},
	},
	Tier:       enum.TierStandard,
	UnitAmount: 100,
	StartUTC:   chrono.TimeNow(),
	EndUTC:     chrono.TimeFrom(time.Now().AddDate(0, 1, 0)),
	Created:    0,
}

var MockStripeStdYearPrice = StripePrice{
	IsFromStripe: false,
	ID:           "price_1IM2nFBzTK0hABgJiIDeDIox",
	Active:       true,
	Currency:     "gbp",
	Kind:         KindRecurring,
	LiveMode:     false,
	Nickname:     "Standard Edition/Year",
	ProductID:    "",
	PeriodCount: ColumnYearMonthDay{
		YearMonthDay: dt.YearMonthDay{
			Years:  1,
			Months: 0,
			Days:   0,
		},
	},
	Tier:       enum.TierStandard,
	UnitAmount: 3900,
	StartUTC:   chrono.Time{},
	EndUTC:     chrono.Time{},
	Created:    0,
}

var MockStripeStdYearCoupons = []StripeCoupon{
	{
		IsFromStripe: false,
		ID:           "",
		AmountOff:    500,
		Created:      0,
		Currency:     "gbp",
		Duration:     null.String{},
		LiveMode:     false,
		Name:         "",
		RedeemBy:     0,
		StripeCouponMeta: StripeCouponMeta{
			PriceID: null.StringFrom(MockStripeStdYearPrice.ID),
			TimeSlot: dt.TimeSlot{
				StartUTC: chrono.Time{},
				EndUTC:   chrono.Time{},
			},
		},
		Status: "",
	},
}

var MockStripeStdMonthPrice = StripePrice{
	IsFromStripe: false,
	ID:           "price_1IM2mgBzTK0hABgJVH8o9Sjm",
	Active:       false,
	Currency:     "gbp",
	Kind:         KindRecurring,
	LiveMode:     false,
	Nickname:     "Standard Edition/Month",
	ProductID:    "",
	PeriodCount: ColumnYearMonthDay{
		YearMonthDay: dt.YearMonthDay{
			Years:  0,
			Months: 1,
			Days:   0,
		},
	},
	Tier:       enum.TierStandard,
	UnitAmount: 499,
	StartUTC:   chrono.Time{},
	EndUTC:     chrono.Time{},
	Created:    0,
}

var MockStripePrmPrice = StripePrice{
	IsFromStripe: false,
	ID:           "plan_FOde0uAr0V4WmT",
	Active:       true,
	Currency:     "gbp",
	Kind:         KindRecurring,
	LiveMode:     false,
	Nickname:     "Premium Edition",
	ProductID:    "",
	PeriodCount: ColumnYearMonthDay{
		YearMonthDay: dt.YearMonthDay{
			Years:  1,
			Months: 0,
			Days:   0,
		},
	},
	Tier:       enum.TierPremium,
	UnitAmount: 23800,
	StartUTC:   chrono.Time{},
	EndUTC:     chrono.Time{},
	Created:    0,
}

func MockRandomStripePrice() StripePrice {
	return StripePrice{
		IsFromStripe: false,
		ID:           faker.StripePriceID(),
		Active:       true,
		Currency:     "gbp",
		Kind:         KindRecurring,
		LiveMode:     false,
		Nickname:     "A mocking price",
		ProductID:    faker.StripeProductID(),
		PeriodCount: ColumnYearMonthDay{
			dt.YearMonthDay{
				Years:  1,
				Months: 0,
				Days:   0,
			},
		},
		Tier:       enum.TierStandard,
		UnitAmount: 3999,
		StartUTC:   chrono.Time{},
		EndUTC:     chrono.Time{},
		Created:    time.Now().Unix(),
	}
}

func (p StripePrice) MockRandomCoupon() StripeCoupon {
	return mockRandomCouponOfPrice(p.ID)
}

func (p StripePrice) MockRandomCouponN(n int) []StripeCoupon {
	var list = make([]StripeCoupon, 0)

	for i := 0; i < n; i++ {
		list = append(list, mockRandomCouponOfPrice(p.ID))
	}

	return list
}

func mockRandomCouponOfPrice(priceId string) StripeCoupon {
	return StripeCoupon{
		IsFromStripe: false,
		ID:           faker.StripeCouponID(),
		AmountOff:    100,
		Created:      0,
		Currency:     "gbp",
		Duration:     null.String{},
		LiveMode:     false,
		Name:         "",
		RedeemBy:     0,
		StripeCouponMeta: StripeCouponMeta{
			PriceID: null.StringFrom(priceId),
			TimeSlot: dt.TimeSlot{
				StartUTC: chrono.TimeUTCNow(),
				EndUTC:   chrono.TimeUTCFrom(time.Now().AddDate(0, 0, 7)),
			},
		},
		Status:     DiscountStatusActive,
		UpdatedUTC: chrono.TimeNow(),
	}
}

func MockRandomStripeCoupon() StripeCoupon {
	return mockRandomCouponOfPrice(faker.StripePriceID())
}

func MockRandomCouponList(n int) []StripeCoupon {
	var list = make([]StripeCoupon, 0)

	priceID := faker.StripePriceID()

	for i := 0; i < n; i++ {
		list = append(list, mockRandomCouponOfPrice(priceID))
	}

	return list
}
