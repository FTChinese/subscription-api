//go:build !production
// +build !production

package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"time"
)

var MockPriceStdIntro = Price{
	Active:   true,
	Created:  1636704870,
	Currency: "gbp",
	ID:       "price_1Juuu2BzTK0hABgJTXiK4NTt",
	LiveMode: false,
	Metadata: PriceMetadata{
		Tier:         enum.TierStandard,
		Introductory: true,
	},
	Nickname:   "Introductory Offer",
	Product:    "prod_FOde1wE4ZTRMcD",
	Recurring:  PriceRecurring{},
	Type:       stripe.PriceTypeOneTime,
	UnitAmount: 10,
	PriceMetadata: PriceMetadata{
		Introductory: true,
		PeriodCount: dt.YearMonthDay{
			Years:  0,
			Months: 1,
			Days:   0,
		},
		Tier:     enum.TierStandard,
		StartUTC: null.String{},
		EndUTC:   null.String{},
	},
}

var MockPriceStdYear = Price{
	Active:   true,
	Created:  1613617385,
	Currency: "gbp",
	ID:       "price_1IM2nFBzTK0hABgJiIDeDIox",
	LiveMode: false,
	Metadata: PriceMetadata{
		Tier:         enum.TierStandard,
		Introductory: false,
	},
	Nickname: "Annual Price",
	Product:  "prod_FOde1wE4ZTRMcD",
	Recurring: PriceRecurring{
		Interval:      stripe.PriceRecurringIntervalYear,
		IntervalCount: 1,
		UsageType:     stripe.PriceRecurringUsageTypeLicensed,
	},
	Type:       stripe.PriceTypeRecurring,
	UnitAmount: 3999,
	PriceMetadata: PriceMetadata{
		Introductory: false,
		PeriodCount: dt.YearMonthDay{
			Years:  1,
			Months: 0,
			Days:   0,
		},
		Tier:     enum.TierStandard,
		StartUTC: null.String{},
		EndUTC:   null.String{},
	},
}

var MockPriceStdMonth = Price{
	Active:   true,
	Created:  1613617350,
	Currency: "gbp",
	ID:       "price_1IM2mgBzTK0hABgJVH8o9Sjm",
	LiveMode: false,
	Metadata: PriceMetadata{
		Tier:         enum.TierStandard,
		Introductory: false,
	},
	Nickname: "Monthly Price",
	Product:  "prod_FOde1wE4ZTRMcD",
	Recurring: PriceRecurring{
		Interval:      stripe.PriceRecurringIntervalMonth,
		IntervalCount: 1,
		UsageType:     stripe.PriceRecurringUsageTypeLicensed,
	},
	Type:       stripe.PriceTypeRecurring,
	UnitAmount: 499,
	PriceMetadata: PriceMetadata{
		Introductory: false,
		PeriodCount: dt.YearMonthDay{
			Years:  0,
			Months: 1,
			Days:   0,
		},
		Tier:     enum.TierStandard,
		StartUTC: null.String{},
		EndUTC:   null.String{},
	},
}

var MockPricePrmYear = Price{
	Active:   true,
	Created:  1562567431,
	Currency: "gbp",
	ID:       "plan_FOde0uAr0V4WmT",
	LiveMode: false,
	Metadata: PriceMetadata{
		Tier:         enum.TierPremium,
		Introductory: false,
	},
	Nickname: "Premium Yearly Price",
	Product:  "prod_FOdd1iNT29BIGq",
	Recurring: PriceRecurring{
		Interval:      stripe.PriceRecurringIntervalYear,
		IntervalCount: 1,
		UsageType:     stripe.PriceRecurringUsageTypeLicensed,
	},
	Type:       stripe.PriceTypeRecurring,
	UnitAmount: 23800,
	PriceMetadata: PriceMetadata{
		Introductory: false,
		PeriodCount: dt.YearMonthDay{
			Years:  1,
			Months: 0,
			Days:   0,
		},
		Tier:     enum.TierPremium,
		StartUTC: null.String{},
		EndUTC:   null.String{},
	},
}

func MockNewSubs() Subs {

	subs, err := NewSubs(faker.MustGenStripeSubs(), ids.UserIDs{
		CompoundID: "",
		FtcID:      null.StringFrom(uuid.New().String()),
		UnionID:    null.String{},
	}.MustNormalize())

	if err != nil {
		panic(err)
	}

	return subs
}

type MockSubsBuilder struct {
	ftcID   string
	edition PriceEdition
	status  enum.SubsStatus
}

func NewMockSubsBuilder(ftcID string) MockSubsBuilder {
	if ftcID == "" {
		ftcID = uuid.New().String()
	}

	return MockSubsBuilder{
		ftcID:   ftcID,
		edition: PriceEditionStore.MustFindByEdition(price.StdYearEdition, false),
		status:  enum.SubsStatusActive,
	}
}

func (b MockSubsBuilder) WithEdition(e price.Edition) MockSubsBuilder {
	b.edition = PriceEditionStore.MustFindByEdition(e, false)
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
	end := dt.NewTimeRange(start).WithCycle(b.edition.Cycle).End
	canceled := time.Time{}

	if b.status == enum.SubsStatusCanceled {
		end = time.Now().AddDate(0, 0, -1)
		start = dt.NewTimeRange(end).WithCycleN(b.edition.Cycle, -1).End
		canceled = end
	}

	return Subs{
		ID:                   faker.GenStripeSubID(),
		Edition:              b.edition.Edition,
		WillCancelAtUtc:      chrono.Time{},
		CancelAtPeriodEnd:    false,
		CanceledUTC:          chrono.TimeFrom(canceled), // Set it for automatic cancel.
		CurrentPeriodEnd:     chrono.TimeFrom(end),
		CurrentPeriodStart:   chrono.TimeFrom(start),
		CustomerID:           faker.GenCustomerID(),
		DefaultPaymentMethod: null.String{},
		SubsItem: SubsItem{
			ItemID:  faker.GenStripeItemID(),
			PriceID: b.edition.PriceID,
		},
		LatestInvoiceID: faker.GenInvoiceID(),
		LiveMode:        false,
		StartDateUTC:    chrono.TimeNow(),
		EndedUTC:        chrono.Time{},
		CreatedUTC:      chrono.TimeNow(),
		UpdatedUTC:      chrono.TimeNow(),
		Status:          b.status,
		FtcUserID:       null.StringFrom(b.ftcID),
		PaymentIntent: PaymentIntent{
			ID:                 faker.GenPaymentIntentID(),
			Amount:             0,
			CanceledAtUTC:      chrono.Time{},
			CancellationReason: "",
			ClientSecret:       null.String{},
			CreatedUtc:         chrono.Time{},
			Currency:           "",
			CustomerID:         "",
			InvoiceID:          "",
			LiveMode:           false,
			PaymentMethodID:    "",
			Status:             "",
		},
	}
}
