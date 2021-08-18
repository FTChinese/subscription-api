package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"time"
)

var FtcOffers = map[Edition][]Discount{
	StdYearEdition: {
		{
			DiscID:   null.StringFrom("2021.8.31-retention-offer-standard-year"),
			PriceOff: null.FloatFrom(149),
			Percent:  null.IntFrom(50),
			DateTimePeriod: dt.DateTimePeriod{
				StartUTC: chrono.TimeNow(),
				EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 3)),
			},
			Description: null.StringFrom("限时特惠 续订享5折"),
			Kind:        OfferKindRetention,
		},
		{
			DiscID:         null.StringFrom("retention-offer-standard-year"),
			PriceOff:       null.FloatFrom(80),
			Percent:        null.IntFrom(75),
			DateTimePeriod: dt.DateTimePeriod{},
			Description:    null.StringFrom("现在续订享75折优惠"),
			Kind:           OfferKindRetention,
		},
		{
			DiscID:         null.StringFrom("win-back-offer-standard-year"),
			PriceOff:       null.FloatFrom(40),
			Percent:        null.IntFrom(85),
			DateTimePeriod: dt.DateTimePeriod{},
			Description:    null.StringFrom("现在购买享85折优惠"),
			Kind:           OfferKindWinBack,
		},
	},
	PremiumEdition: {
		{
			DiscID:   null.StringFrom("2021.8.31-retention-offer-premium-year"),
			PriceOff: null.FloatFrom(999),
			Percent:  null.IntFrom(50),
			DateTimePeriod: dt.DateTimePeriod{
				StartUTC: chrono.TimeNow(),
				EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 3)),
			},
			Description: null.StringFrom("限时特惠 续订享5折"),
			Kind:        OfferKindRetention,
		},
		{
			DiscID:         null.StringFrom("retention-offer-premium-year"),
			PriceOff:       null.FloatFrom(500),
			Percent:        null.IntFrom(75),
			DateTimePeriod: dt.DateTimePeriod{},
			Description:    null.StringFrom("现在续订享75折优惠"),
			Kind:           OfferKindRetention,
		},
		{
			DiscID:         null.StringFrom("win-back-offer-premium-year"),
			PriceOff:       null.FloatFrom(300),
			Percent:        null.IntFrom(85),
			DateTimePeriod: dt.DateTimePeriod{},
			Description:    null.StringFrom("现在购买享85折优惠"),
			Kind:           OfferKindWinBack,
		},
	},
}
