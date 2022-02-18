//go:build !production
// +build !production

package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	sdk "github.com/stripe/stripe-go/v72"
	"time"
)

func (p Persona) StripeCustomer() stripe.Customer {
	return stripe.Customer{
		IsFromStripe:           false,
		ID:                     faker.GenStripeCusID(),
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
		ID:                   faker.GenInvoiceID(),
		AutoAdvance:          true,
		ChargeID:             "",
		CollectionMethod:     sdk.InvoiceCollectionMethodChargeAutomatically,
		Currency:             "gbp",
		CustomerID:           faker.GenStripeCusID(),
		DefaultPaymentMethod: null.String{},
		HostedInvoiceURL:     null.String{},
		LiveMode:             false,
		Paid:                 true,
		PaymentIntentID:      "",
		PeriodEndUTC:         chrono.TimeNow(),
		PeriodStartUTC:       chrono.TimeNow(),
		ReceiptNumber:        "",
		Status:               sdk.InvoiceStatusOpen,
		SubscriptionID:       null.StringFrom(faker.GenStripeSubID()),
		Total:                0,
		Created:              time.Now().Unix(),
	}
}

func StripeSetupIntent() stripe.SetupIntent {
	return stripe.SetupIntent{
		ID:                 faker.GenSetupIntentID(),
		CancellationReason: "",
		ClientSecret:       rand.String(40),
		Created:            time.Now().Unix(),
		CustomerID:         faker.GenStripeCusID(),
		LiveMode:           false,
		NextAction:         stripe.SetupIntentNextActionJSON{},
		PaymentMethodID:    null.String{},
		PaymentMethodTypes: nil,
		Status:             "",
		Usage:              sdk.SetupIntentUsageOffSession,
	}
}

func StripePaymentIntent() stripe.PaymentIntent {
	return stripe.PaymentIntent{
		ID:                 faker.GenPaymentIntentID(),
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
		NextAction:         stripe.NextActionJSON{},
		PaymentMethodID:    "",
		PaymentMethodTypes: nil,
		ReceiptEmail:       "",
		SetupFutureUsage:   "",
		Status:             "",
	}
}

func StripePaymentMethod() stripe.PaymentMethod {
	return stripe.PaymentMethod{
		ID:         faker.GenPaymentMethodID(),
		CustomerID: faker.GenStripeCusID(),
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
		price:  stripe.MockPriceStdYear,
		status: enum.SubsStatusActive,
	}
}

type StripeSubsBuilder struct {
	ftcID  string
	price  stripe.Price
	status enum.SubsStatus
}

func (b StripeSubsBuilder) WithEdition(e price.Edition) StripeSubsBuilder {
	return b
}

func (b StripeSubsBuilder) WithPrice(p stripe.Price) StripeSubsBuilder {
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
	end := dt.NewTimeRange(start).WithPeriod(b.price.PeriodCount).End
	canceled := time.Time{}

	subsID := faker.GenStripeSubID()

	return stripe.Subs{
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
		Items: []stripe.SubsItem{
			{
				ID: faker.GenStripeItemID(),
				Price: stripe.PriceJSON{
					Price: b.price,
				},
				Created:        time.Now().Unix(),
				Quantity:       1,
				SubscriptionID: subsID,
			},
		},
		LatestInvoiceID: faker.GenInvoiceID(),
		LatestInvoice:   stripe.Invoice{},
		LiveMode:        false,
		PaymentIntentID: null.StringFrom(faker.GenPaymentIntentID()),
		PaymentIntent:   stripe.PaymentIntent{},
		StartDateUTC:    chrono.TimeNow(),
		Status:          b.status,
		Created:         time.Now().Unix(),
	}
}
