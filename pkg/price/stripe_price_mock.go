//go:build !production
// +build !production

package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"time"
)

func MockNewStripePrice() StripePrice {
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
		PeriodCount: dt.YearMonthDay{
			Months: 0,
			Days:   0,
		},
		Tier:       enum.TierStandard,
		UnitAmount: 3999,
		StartUTC:   null.StringFrom(chrono.TimeNow().Format(time.RFC3339)),
		EndUTC:     null.StringFrom(chrono.TimeNow().AddDate(1, 0, 0).Format(time.RFC3339)),
		Created:    time.Now().Unix(),
	}
}
