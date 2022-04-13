//go:build !production

package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"time"
)

var MockStripeStdIntroPrice = StripePrice{
	IsFromStripe:   false,
	ID:             "price_1Juuu2BzTK0hABgJTXiK4NTt",
	Active:         true,
	Currency:       "gbp",
	IsIntroductory: true,
	Kind:           KindOneTime,
	LiveMode:       false,
	Nickname:       "Trial period",
	ProductID:      "",
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
	IsFromStripe:   false,
	ID:             "price_1IM2nFBzTK0hABgJiIDeDIox",
	Active:         true,
	Currency:       "gbp",
	IsIntroductory: false,
	Kind:           KindRecurring,
	LiveMode:       false,
	Nickname:       "Standard Edition/Year",
	ProductID:      "",
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
		Duration:     "once",
		LiveMode:     false,
		Name:         "",
		RedeemBy:     0,
		StripeCouponMeta: StripeCouponMeta{
			PriceID:  null.StringFrom(MockStripeStdYearPrice.ID),
			StartUTC: chrono.Time{},
			EndUTC:   chrono.Time{},
		},
		Status: "",
	},
}

var MockStripeStdMonthPrice = StripePrice{
	IsFromStripe:   false,
	ID:             "price_1IM2mgBzTK0hABgJVH8o9Sjm",
	Active:         false,
	Currency:       "gbp",
	IsIntroductory: false,
	Kind:           KindRecurring,
	LiveMode:       false,
	Nickname:       "Standard Edition/Month",
	ProductID:      "",
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
	IsFromStripe:   false,
	ID:             "plan_FOde0uAr0V4WmT",
	Active:         true,
	Currency:       "gbp",
	IsIntroductory: false,
	Kind:           KindRecurring,
	LiveMode:       false,
	Nickname:       "Premium Edition",
	ProductID:      "",
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
		IsFromStripe:   false,
		ID:             faker.GenStripePriceID(),
		Active:         true,
		Currency:       "gbp",
		IsIntroductory: false,
		Kind:           KindRecurring,
		LiveMode:       false,
		Nickname:       "A mocking price",
		ProductID:      faker.GenStripeProductID(),
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

func MockRandomStripeCoupon() StripeCoupon {
	return StripeCoupon{
		IsFromStripe: false,
		ID:           faker.GenStripePriceID(),
		AmountOff:    100,
		Created:      0,
		Currency:     "",
		Duration:     "gbp",
		LiveMode:     false,
		Name:         "",
		RedeemBy:     0,
		StripeCouponMeta: StripeCouponMeta{
			PriceID:  null.StringFrom(faker.GenStripePriceID()),
			StartUTC: chrono.TimeNow(),
			EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 7)),
		},
		Status: DiscountStatusActive,
	}
}
