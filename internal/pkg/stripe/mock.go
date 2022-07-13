//go:build !production

package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
	sdk "github.com/stripe/stripe-go/v72"
	"time"
)

func MockRandomDiscount() Discount {
	return Discount{
		IsFromStripe:    false,
		ID:              faker.StripeDiscountID(),
		Coupon:          CouponColumn{price.MockRandomStripeCoupon()},
		CustomerID:      faker.StripeCustomerID(),
		End:             null.Int{},
		InvoiceID:       null.String{},
		InvoiceItemID:   null.String{},
		PromotionCodeID: null.String{},
		Start:           0,
		SubsID:          null.String{},
	}
}

type MockSubsBuilder struct {
	ftcID    string
	price    price.StripePrice
	status   enum.SubsStatus
	discount Discount
}

func NewMockSubsBuilder(ftcID string) MockSubsBuilder {
	return MockSubsBuilder{
		ftcID:  ftcID,
		price:  price.MockStripeStdYearPrice,
		status: enum.SubsStatusActive,
	}
}

func (b MockSubsBuilder) WithPrice(p price.StripePrice) MockSubsBuilder {
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

func (b MockSubsBuilder) WithDiscount() MockSubsBuilder {
	b.discount = MockRandomDiscount()
	return b
}

func (b MockSubsBuilder) Build() Subs {
	start := time.Now()
	end := dt.NewSlotBuilder(start).WithPeriod(b.price.PeriodCount.YearMonthDay).End
	canceled := time.Time{}

	subsID := faker.StripeSubsID()

	return Subs{
		ID:                     subsID,
		Edition:                b.price.Edition(),
		WillCancelAtUtc:        chrono.Time{},
		CancelAtPeriodEnd:      false,
		CanceledUTC:            chrono.TimeFrom(canceled), // Set it for automatic cancel.
		CurrentPeriodEnd:       chrono.TimeFrom(end),
		CurrentPeriodStart:     chrono.TimeFrom(start),
		CustomerID:             faker.StripeCustomerID(),
		DefaultPaymentMethodID: null.String{},
		Discount:               DiscountColumn{b.discount},
		EndedUTC:               chrono.Time{},
		FtcUserID:              null.StringFrom(b.ftcID),
		Items: []SubsItem{
			{
				ID:             faker.StripeSubsItemID(),
				Price:          b.price,
				Created:        time.Now().Unix(),
				Quantity:       1,
				SubscriptionID: subsID,
			},
		},
		LatestInvoiceID: faker.StripeInvoiceID(),
		LatestInvoice:   Invoice{},
		LiveMode:        false,
		PaymentIntentID: null.StringFrom(faker.StripePaymentIntentID()),
		PaymentIntent:   PaymentIntent{},
		StartDateUTC:    chrono.TimeNow(),
		Status:          b.status,
		Created:         time.Now().Unix(),
	}
}

func MockPaymentMethod() PaymentMethod {
	return PaymentMethod{
		ID:         faker.StripePaymentMethodID(),
		CustomerID: faker.StripeCustomerID(),
		Kind:       sdk.PaymentMethodTypeCard,
		Card: PaymentMethodCard{
			Brand:             "visa",
			Country:           "us",
			ExpMonth:          9,
			ExpYear:           23,
			Fingerprint:       "",
			Funding:           "",
			Last4:             "1234",
			Network:           sdk.PaymentMethodCardNetworks{},
			ThreeDSecureUsage: sdk.PaymentMethodCardThreeDSecureUsage{},
		},
		Created:  time.Now().Unix(),
		LiveMode: false,
	}
}

func MockInvoice() Invoice {
	return Invoice{
		ID:                   faker.StripeInvoiceID(),
		AutoAdvance:          true,
		ChargeID:             "",
		CollectionMethod:     InvoiceCollectionMethod{},
		Currency:             "gbp",
		CustomerID:           faker.StripeCustomerID(),
		DefaultPaymentMethod: null.StringFrom("payment-method-id"),
		Discounts:            []string{"discount-1", "discount-b"},
		HostedInvoiceURL:     gofakeit.URL(),
		LiveMode:             false,
		Paid:                 true,
		PaymentIntentID:      "",
		PeriodEndUTC:         chrono.TimeNow(),
		PeriodStartUTC:       chrono.TimeNow(),
		ReceiptNumber:        "",
		Status:               InvoiceStatus{},
		SubscriptionID:       null.StringFrom(faker.StripeSubsID()),
		Total:                0,
		Created:              time.Now().Unix(),
	}
}
