package price

import (
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
)

var FtcOffers = map[Edition][]Discount{
	StdYearEdition: {
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
			Description:    null.StringFrom("再次订阅享85折优惠"),
			Kind:           OfferKindWinBack,
		},
	},
	PremiumEdition: {
		{
			DiscID:         null.StringFrom("retention-offer-premium-year"),
			PriceOff:       null.FloatFrom(500),
			Percent:        null.IntFrom(75),
			DateTimePeriod: dt.DateTimePeriod{},
			Description:    null.StringFrom("现在升级或续订享75折优惠"),
			Kind:           OfferKindRetention,
		},
		{
			DiscID:         null.StringFrom("win-back-offer-premium-year"),
			PriceOff:       null.FloatFrom(300),
			Percent:        null.IntFrom(85),
			DateTimePeriod: dt.DateTimePeriod{},
			Description:    null.StringFrom("再次订阅享85折优惠"),
			Kind:           OfferKindWinBack,
		},
	},
}
