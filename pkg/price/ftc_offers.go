package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"time"
)

// 2021-08-22T16:00:00Z
var (
	//2021-08-22T16:00:00Z
	retentionHalfStart = time.Date(2021, 8, 22, 16, 0, 0, 0, time.UTC)
	// 2021-09-02T16:00:00Z
	retentionHalfEnd = time.Date(2021, 9, 2, 16, 0, 0, 0, time.UTC)
)

var FtcOffers = map[Edition][]Discount{
	StdYearEdition: {
		// Discount for 2021-08.
		{
			DiscID:   null.StringFrom("21.8.31-retention-standard-year"),
			PriceOff: null.FloatFrom(148),
			Percent:  null.IntFrom(50),
			DateTimePeriod: dt.DateTimePeriod{
				// 2021-08-22T16:00:00Z
				// 2021-09-02T16:00:00Z
				StartUTC: chrono.TimeFrom(time.Now()),
				EndUTC:   chrono.TimeFrom(retentionHalfEnd),
			},
			Description: null.StringFrom("限时特惠 续订享5折"),
			Kind:        OfferKindRetention,
		},
		// Retention that is in effect forever.
		{
			DiscID:         null.StringFrom("retention-offer-standard-year"),
			PriceOff:       null.FloatFrom(80),
			Percent:        null.IntFrom(75),
			DateTimePeriod: dt.DateTimePeriod{},
			Description:    null.StringFrom("现在续订享75折优惠"),
			Kind:           OfferKindRetention,
		},
		// Winback that is in effect forever.
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
		// Discount for 2021-08.
		{
			DiscID:   null.StringFrom("21.8.31-retention-premium-year"),
			PriceOff: null.FloatFrom(998),
			Percent:  null.IntFrom(50),
			DateTimePeriod: dt.DateTimePeriod{
				StartUTC: chrono.TimeFrom(time.Now()),
				EndUTC:   chrono.TimeFrom(retentionHalfEnd),
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
