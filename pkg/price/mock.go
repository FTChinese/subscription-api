// +build !production

package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"time"
)

var PriceStdYear = FtcPrice{
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
			DiscID:   null.StringFrom("dsc_F7gEwjaF3OsR"),
			PriceOff: null.FloatFrom(130),
			Percent:  null.Int{},
			DateTimePeriod: dt.DateTimePeriod{
				StartUTC: chrono.TimeFrom(time.Date(2021, 2, 1, 4, 0, 0, 0, time.UTC)),
				EndUTC:   chrono.TimeFrom(time.Date(2021, 2, 7, 16, 0, 0, 0, time.UTC)),
			},
			Description: null.String{},
		},
		FtcOffers[StdYearEdition][0],
		FtcOffers[StdYearEdition][1],
	},
}

var PriceStdMonth = FtcPrice{
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
}

var PricePrm = FtcPrice{
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
			DiscID:   null.StringFrom("dsc_7VQy0Zvifacq"),
			PriceOff: null.FloatFrom(300),
			Percent:  null.Int{},
			DateTimePeriod: dt.DateTimePeriod{
				StartUTC: chrono.TimeFrom(time.Date(2021, 2, 1, 4, 0, 0, 0, time.UTC)),
				EndUTC:   chrono.TimeFrom(time.Date(2021, 2, 7, 16, 0, 0, 0, time.UTC)),
			},
			Description: null.StringFrom("限时促销"),
			Kind:        OfferKindPromotion,
		},
		FtcOffers[PremiumEdition][0],
		FtcOffers[PremiumEdition][1],
	},
}
