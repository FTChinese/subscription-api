//go:build !production
// +build !production

package stripe

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

var MockPriceStdIntro = Price{
	ID:             "price_1Juuu2BzTK0hABgJTXiK4NTt",
	Active:         true,
	Currency:       "gbp",
	IsIntroductory: true,
	Kind:           price.KindOneTime,
	LiveMode:       false,
	Nickname:       "Introductory Offer",
	ProductID:      "prod_FOde1wE4ZTRMcD",
	PeriodCount: dt.YearMonthDay{
		Years:  0,
		Months: 1,
		Days:   0,
	},
	Tier:       enum.TierStandard,
	UnitAmount: 10,
	StartUTC:   null.String{},
	EndUTC:     null.String{},
	Created:    1636704870,
}

var MockPriceStdYear = Price{
	ID:             "price_1IM2nFBzTK0hABgJiIDeDIox",
	Active:         true,
	Currency:       "gbp",
	IsIntroductory: false,
	Kind:           price.KindRecurring,
	LiveMode:       false,
	Nickname:       "Regular Yearly Charge",
	ProductID:      "prod_FOde1wE4ZTRMcD",
	PeriodCount: dt.YearMonthDay{
		Years:  1,
		Months: 0,
		Days:   0,
	},
	Tier:       enum.TierStandard,
	UnitAmount: 3999,
	StartUTC:   null.String{},
	EndUTC:     null.String{},
	Created:    1613617385,
}

var MockPriceStdMonth = Price{
	ID:             "price_1IM2mgBzTK0hABgJVH8o9Sjm",
	Active:         true,
	Currency:       "gbp",
	IsIntroductory: false,
	Kind:           price.KindRecurring,
	LiveMode:       false,
	Nickname:       "Regular Monthly Charge",
	ProductID:      "prod_FOde1wE4ZTRMcD",
	PeriodCount: dt.YearMonthDay{
		Years:  0,
		Months: 1,
		Days:   0,
	},
	Tier:       enum.TierStandard,
	UnitAmount: 499,
	StartUTC:   null.String{},
	EndUTC:     null.String{},
	Created:    1613617350,
}

var MockPricePrmYear = Price{
	ID:             "plan_FOde0uAr0V4WmT",
	Active:         true,
	Currency:       "gbp",
	IsIntroductory: false,
	Kind:           price.KindRecurring,
	LiveMode:       false,
	Nickname:       "Premium Yearly Price",
	ProductID:      "prod_FOdd1iNT29BIGq",
	PeriodCount: dt.YearMonthDay{
		Years:  1,
		Months: 0,
		Days:   0,
	},
	Tier:       enum.TierPremium,
	UnitAmount: 23800,
	StartUTC:   null.String{},
	EndUTC:     null.String{},
	Created:    1562567431,
}
