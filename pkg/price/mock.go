// +build !production

package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"time"
)

var MockPriceStdYear = FtcPrice{
	Price: Price{
		ID: "plan_MynUQDQY1TSQ",
		Edition: Edition{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleYear,
		},
		Active:     true,
		Currency:   CurrencyCNY,
		LiveMode:   true,
		Nickname:   null.String{},
		ProductID:  "prod_zjWdiTUpDN8l",
		Source:     SourceFTC,
		UnitAmount: 298,
	},
	Offers: []Discount{
		{
			ID: "dsc_F7gEwjaF3OsR",
			DiscountParams: DiscountParams{
				CreatedBy:   "anonymous",
				Description: null.String{},
				Kind:        "",
				Percent:     null.Int{},
				DateTimePeriod: dt.DateTimePeriod{
					StartUTC: chrono.TimeFrom(time.Date(2021, 2, 1, 4, 0, 0, 0, time.UTC)),
					EndUTC:   chrono.TimeFrom(time.Date(2021, 2, 7, 16, 0, 0, 0, time.UTC)),
				},
				PriceOff:  null.FloatFrom(130),
				PriceID:   "plan_MynUQDQY1TSQ",
				Recurring: false,
			},
			LiveMode:   false,
			Status:     "",
			CreatedUTC: chrono.Time{},
		},
		{
			ID: "retention-offer-standard-year",
			DiscountParams: DiscountParams{
				PriceOff:       null.FloatFrom(80),
				Percent:        null.IntFrom(75),
				DateTimePeriod: dt.DateTimePeriod{},
				Description:    null.StringFrom("现在续订享75折优惠"),
				Kind:           OfferKindRetention,
			},
			Status:     "DiscountStatusActive",
			CreatedUTC: chrono.Time{},
		},
		// Winback that is in effect forever.
		{
			ID: "win-back-offer-standard-year",
			DiscountParams: DiscountParams{
				PriceOff:       null.FloatFrom(40),
				Percent:        null.IntFrom(85),
				DateTimePeriod: dt.DateTimePeriod{},
				Description:    null.StringFrom("重新购买会员享85折优惠"),
				Kind:           OfferKindWinBack,
			},
			Status:     DiscountStatusActive,
			CreatedUTC: chrono.Time{},
		},
	},
}

var MockPriceStdMonth = FtcPrice{
	Price: Price{
		ID: "plan_1Uz4hrLy3Mzy",
		Edition: Edition{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleMonth,
		},
		Active:     true,
		Currency:   CurrencyCNY,
		LiveMode:   true,
		Nickname:   null.String{},
		ProductID:  "prod_zjWdiTUpDN8l",
		Source:     SourceFTC,
		UnitAmount: 28,
	},
	Offers: []Discount{
		{
			ID: "intro-offer-std-month",
			DiscountParams: DiscountParams{
				PriceOff: null.FloatFrom(34),
				Percent:  null.Int{},
				DateTimePeriod: dt.DateTimePeriod{
					StartUTC: chrono.TimeNow(),
					EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 7)),
				},
				Description: null.StringFrom("新会员订阅仅需1元"),
				Kind:        OfferKindIntroductory,
			},
			Status:     DiscountStatusActive,
			CreatedUTC: chrono.TimeNow(),
		},
	},
}

var MockPricePrm = FtcPrice{
	Price: Price{
		ID: "plan_vRUzRQ3aglea",
		Edition: Edition{
			Tier:  enum.TierPremium,
			Cycle: enum.CycleYear,
		},
		Active:     true,
		Currency:   CurrencyCNY,
		LiveMode:   true,
		Nickname:   null.String{},
		ProductID:  "prod_IaoK5SbK79g8",
		Source:     SourceFTC,
		UnitAmount: 1998,
	},
	Offers: []Discount{
		{
			ID: "dsc_7VQy0Zvifacq",
			DiscountParams: DiscountParams{
				PriceOff: null.FloatFrom(300),
				Percent:  null.Int{},
				DateTimePeriod: dt.DateTimePeriod{
					StartUTC: chrono.TimeFrom(time.Date(2021, 2, 1, 4, 0, 0, 0, time.UTC)),
					EndUTC:   chrono.TimeFrom(time.Date(2021, 2, 7, 16, 0, 0, 0, time.UTC)),
				},
				Description: null.StringFrom("限时促销"),
				Kind:        OfferKindPromotion,
			},
		},
		{
			ID: "retention-offer-premium-year",
			DiscountParams: DiscountParams{
				PriceOff:       null.FloatFrom(500),
				Percent:        null.IntFrom(75),
				DateTimePeriod: dt.DateTimePeriod{},
				Description:    null.StringFrom("现在续订享75折优惠"),
				Kind:           OfferKindRetention,
			},
			Status:     DiscountStatusActive,
			CreatedUTC: chrono.Time{},
		},
		{
			ID: "win-back-offer-premium-year",
			DiscountParams: DiscountParams{
				PriceOff:       null.FloatFrom(300),
				Percent:        null.IntFrom(85),
				DateTimePeriod: dt.DateTimePeriod{},
				Description:    null.StringFrom("重新购买会员享85折优惠"),
				Kind:           OfferKindWinBack,
			},
			Status:     DiscountStatusActive,
			CreatedUTC: chrono.Time{},
		},
	},
}
