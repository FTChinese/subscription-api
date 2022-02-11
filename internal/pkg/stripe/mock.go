//go:build !production
// +build !production

package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"time"
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

type MockSubsBuilder struct {
	ftcID  string
	price  Price
	status enum.SubsStatus
}

func NewMockSubsBuilder(ftcID string) MockSubsBuilder {
	if ftcID == "" {
		ftcID = uuid.New().String()
	}

	return MockSubsBuilder{
		ftcID:  ftcID,
		price:  MockPriceStdYear,
		status: enum.SubsStatusActive,
	}
}

func (b MockSubsBuilder) WithEdition(e price.Edition) MockSubsBuilder {
	return b
}

func (b MockSubsBuilder) WithPrice(p Price) MockSubsBuilder {
	b.price = p
	return b
}

func (b MockSubsBuilder) WithStatus(s enum.SubsStatus) MockSubsBuilder {
	b.status = s
	return b
}

func (b MockSubsBuilder) WithCanceled() MockSubsBuilder {
	return b.WithStatus(enum.SubsStatusCanceled)
}

func (b MockSubsBuilder) Build() Subs {
	start := time.Now()
	end := dt.NewTimeRange(start).WithPeriod(b.price.PeriodCount).End
	canceled := time.Time{}

	subsID := faker.GenStripeSubID()

	return Subs{
		ID:                     subsID,
		Edition:                b.price.Edition(),
		WillCancelAtUtc:        chrono.Time{},
		CancelAtPeriodEnd:      false,
		CanceledUTC:            chrono.TimeFrom(canceled), // Set it for automatic cancel.
		CurrentPeriodEnd:       chrono.TimeFrom(end),
		CurrentPeriodStart:     chrono.TimeFrom(start),
		CustomerID:             faker.GenStripeCusID(),
		DefaultPaymentMethodID: null.String{},
		EndedUTC:               chrono.Time{},
		FtcUserID:              null.StringFrom(b.ftcID),
		Items: []SubsItem{
			{
				ID: faker.GenStripeItemID(),
				Price: PriceJSON{
					Price: b.price,
				},
				Created:        time.Now().Unix(),
				Quantity:       1,
				SubscriptionID: subsID,
			},
		},
		LatestInvoiceID: faker.GenInvoiceID(),
		LatestInvoice:   Invoice{},
		LiveMode:        false,
		PaymentIntentID: null.String{},
		PaymentIntent:   PaymentIntent{},
		StartDateUTC:    chrono.TimeNow(),
		Status:          b.status,
		Created:         time.Now().Unix(),
	}
}
