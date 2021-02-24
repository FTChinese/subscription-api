// +build !production

package faker

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

var PriceStdYear = price.FtcPrice{
	Original: price.Price{
		ID: "plan_MynUQDQY1TSQ",
		Edition: price.Edition{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleYear,
		},
		Active:     true,
		Currency:   price.CurrencyCNY,
		LiveMode:   true,
		Nickname:   null.String{},
		ProductID:  "prod_zjWdiTUpDN8l",
		Source:     price.SourceFTC,
		UnitAmount: 298,
	},
	PromotionOffer: price.Discount{
		//DiscID:   null.StringFrom("dsc_F7gEwjaF3OsR"),
		//PriceOff: null.FloatFrom(130),
		//Percent:  null.Int{},
		//Period: price.Period{
		//	StartUTC: chrono.TimeNow(),
		//	EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 2)),
		//},
		//Description: null.String{},
	},
}

var PriceStdMonth = price.FtcPrice{
	Original: price.Price{
		ID: "plan_1Uz4hrLy3Mzy",
		Edition: price.Edition{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleMonth,
		},
		Active:     true,
		Currency:   price.CurrencyCNY,
		LiveMode:   true,
		Nickname:   null.String{},
		ProductID:  "prod_zjWdiTUpDN8l",
		Source:     price.SourceFTC,
		UnitAmount: 28,
	},
	PromotionOffer: price.Discount{},
}

var PricePrm = price.FtcPrice{
	Original: price.Price{
		ID: "plan_vRUzRQ3aglea",
		Edition: price.Edition{
			Tier:  enum.TierPremium,
			Cycle: enum.CycleYear,
		},
		Active:     true,
		Currency:   price.CurrencyCNY,
		LiveMode:   true,
		Nickname:   null.String{},
		ProductID:  "prod_IaoK5SbK79g8",
		Source:     price.SourceFTC,
		UnitAmount: 1998,
	},
	PromotionOffer: price.Discount{},
}
