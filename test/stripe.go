//go:build !production

package test

import (
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	sdk "github.com/stripe/stripe-go/v72"
)

func (p Persona) StripeCustomer() stripe.Customer {
	return stripe.Customer{
		IsFromStripe:           false,
		ID:                     faker.StripeCustomerID(),
		FtcID:                  p.FtcID,
		Currency:               "gbp",
		Created:                time.Now().Unix(),
		DefaultSource:          null.String{},
		DefaultPaymentMethodID: null.String{},
		Email:                  p.Email,
		LiveMode:               false,
	}
}

func StripeInvoice() stripe.Invoice {
	return stripe.Invoice{
		ID:                   faker.StripeInvoiceID(),
		AutoAdvance:          true,
		ChargeID:             "",
		CollectionMethod:     stripe.InvoiceCollectionMethod{},
		Currency:             "gbp",
		CustomerID:           faker.StripeCustomerID(),
		DefaultPaymentMethod: null.String{},
		HostedInvoiceURL:     null.String{},
		LiveMode:             false,
		Paid:                 true,
		PaymentIntentID:      "",
		PeriodEndUTC:         chrono.TimeNow(),
		PeriodStartUTC:       chrono.TimeNow(),
		ReceiptNumber:        "",
		Status:               stripe.InvoiceStatus{},
		SubscriptionID:       null.StringFrom(faker.StripeSubsID()),
		Total:                0,
		Created:              time.Now().Unix(),
	}
}

func StripeSetupIntent() stripe.SetupIntent {
	return stripe.SetupIntent{
		ID:                 faker.StripeSetupIntentID(),
		CancellationReason: stripe.SICancelReason{},
		ClientSecret:       rand.String(40),
		Created:            time.Now().Unix(),
		CustomerID:         faker.StripeCustomerID(),
		LiveMode:           false,
		NextAction:         stripe.SINextActionJSON{},
		PaymentMethodID:    null.String{},
		PaymentMethodTypes: nil,
		Status:             stripe.SIStatus{},
		Usage:              stripe.SIUsage{},
	}
}

func StripePaymentIntent() stripe.PaymentIntent {
	return stripe.PaymentIntent{
		ID:                 faker.StripePaymentIntentID(),
		Amount:             299,
		AmountReceived:     299,
		CanceledAt:         0,
		CancellationReason: "",
		ClientSecret:       null.String{},
		Created:            0,
		Currency:           "",
		CustomerID:         "",
		InvoiceID:          "",
		LiveMode:           false,
		NextAction:         stripe.PINextActionJSON{},
		PaymentMethodID:    "",
		PaymentMethodTypes: nil,
		ReceiptEmail:       "",
		SetupFutureUsage:   stripe.SetupFutureUsage{},
		Status:             "",
	}
}

func StripePaymentMethod() stripe.PaymentMethod {
	return stripe.PaymentMethod{
		ID:         faker.StripePaymentMethodID(),
		CustomerID: faker.StripeCustomerID(),
		Kind:       sdk.PaymentMethodTypeCard,
		Card: stripe.PaymentMethodCard{
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

func (p Persona) StripeSubsBuilder() StripeSubsBuilder {
	return StripeSubsBuilder{
		ftcID:  p.FtcID,
		price:  price.MockStripeStdYearPrice,
		status: enum.SubsStatusActive,
	}
}

type StripeSubsBuilder struct {
	ftcID  string
	price  price.StripePrice
	status enum.SubsStatus
}

func (b StripeSubsBuilder) WithPrice(p price.StripePrice) StripeSubsBuilder {
	b.price = p
	return b
}

func (b StripeSubsBuilder) WithStatus(s enum.SubsStatus) StripeSubsBuilder {
	b.status = s
	return b
}

func (b StripeSubsBuilder) WithCanceled() StripeSubsBuilder {
	return b.WithStatus(enum.SubsStatusCanceled)
}

func (b StripeSubsBuilder) Build() stripe.Subs {
	start := time.Now()
	end := dt.NewSlotBuilder(start).WithPeriod(b.price.PeriodCount.YearMonthDay).End
	canceled := time.Time{}

	subsID := faker.StripeSubsID()

	return stripe.Subs{
		ID:                     subsID,
		Edition:                b.price.Edition(),
		WillCancelAtUtc:        chrono.Time{},
		CancelAtPeriodEnd:      false,
		CanceledUTC:            chrono.TimeFrom(canceled), // Set it for automatic cancel.
		CurrentPeriodEnd:       chrono.TimeFrom(end),
		CurrentPeriodStart:     chrono.TimeFrom(start),
		CustomerID:             faker.StripeCustomerID(),
		DefaultPaymentMethodID: null.String{},
		EndedUTC:               chrono.Time{},
		FtcUserID:              null.StringFrom(b.ftcID),
		Items: []stripe.SubsItem{
			{
				ID:             faker.StripeSubsItemID(),
				Price:          b.price,
				Created:        time.Now().Unix(),
				Quantity:       1,
				SubscriptionID: subsID,
			},
		},
		LatestInvoiceID: faker.StripeInvoiceID(),
		LatestInvoice:   stripe.Invoice{},
		LiveMode:        false,
		PaymentIntentID: null.StringFrom(faker.StripePaymentIntentID()),
		PaymentIntent:   stripe.PaymentIntent{},
		StartDateUTC:    chrono.TimeNow(),
		Status:          b.status,
		Created:         time.Now().Unix(),
	}
}
