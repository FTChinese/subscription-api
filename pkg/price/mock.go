//go:build !production
// +build !production

package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"time"
)

var MockEditionStdYear = Edition{
	Tier:  enum.TierStandard,
	Cycle: enum.CycleYear,
}

var MockEditionStdMonth = Edition{
	Tier:  enum.TierStandard,
	Cycle: enum.CycleMonth,
}

var MockEditionPrm = Edition{
	Tier:  enum.TierPremium,
	Cycle: enum.CycleYear,
}

var MockIntroPrice = FtcPrice{
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
	ProductID:     "prod_9xrJdHFq0wmq",
	StripePriceID: "price_1Juuu2BzTK0hABgJTXiK4NTt",
	Title:         null.StringFrom("新会员首次订阅1元/月"),
	UnitAmount:    1,
	StartUTC:      chrono.TimeNow(),
	EndUTC:        chrono.TimeFrom(time.Now().AddDate(0, 0, 7)),
	CreatedUTC:    chrono.TimeNow(),
}

var MockPriceStdYear = FtcPrice{
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
	ProductID:     "prod_9xrJdHFq0wmq",
	StripePriceID: "price_1IM2nFBzTK0hABgJiIDeDIox",
	Title:         null.StringFrom("Standard Annual Edition"),
	UnitAmount:    298,
	StartUTC:      chrono.Time{},
	EndUTC:        chrono.Time{},
	CreatedUTC:    chrono.TimeNow(),
}

var MockStdYearOffers = []Discount{
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
			Percent:      null.Int{},
			PriceOff:     null.FloatFrom(80),
			PriceID:      "price_WHc5ssjh6pqw",
			Recurring:    true,
			ChronoPeriod: dt.ChronoPeriod{},
			CreatedBy:    "",
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
			Percent:      null.Int{},
			PriceOff:     null.FloatFrom(40),
			PriceID:      "price_WHc5ssjh6pqw",
			Recurring:    true,
			ChronoPeriod: dt.ChronoPeriod{},
			CreatedBy:    "",
		},
		CreatedUTC: chrono.Time{},
	},
}

var MockPriceStdMonth = FtcPrice{
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
	ProductID:     "prod_9xrJdHFq0wmq",
	StripePriceID: "price_1IM2mgBzTK0hABgJVH8o9Sjm",
	Title:         null.StringFrom("Standard Monthly Edition"),
	UnitAmount:    28,
	StartUTC:      chrono.Time{},
	EndUTC:        chrono.Time{},
	CreatedUTC:    chrono.TimeNow(),
}

var MockStdMonthOffers []Discount

var MockPricePrm = FtcPrice{
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
	ProductID:     "prod_zSgOTS6DWLmu",
	StripePriceID: "plan_FOde0uAr0V4WmT",
	Title:         null.StringFrom("Premium Annual Edition"),
	UnitAmount:    1998,
	StartUTC:      chrono.Time{},
	EndUTC:        chrono.Time{},
	CreatedUTC:    chrono.TimeNow(),
}

var MockPrmOffers = []Discount{
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
			Percent:      null.Int{},
			PriceOff:     null.FloatFrom(500),
			PriceID:      "price_zsTj2TQ1h3jB",
			Recurring:    true,
			ChronoPeriod: dt.ChronoPeriod{},
			CreatedBy:    "",
		},
		CreatedUTC: chrono.Time{},
	},
}
