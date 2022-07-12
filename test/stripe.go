//go:build !production

package test

import (
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/guregu/null"
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
		DefaultPaymentMethod: null.StringFrom("payment-method-id"),
		Discounts:            []string{"discount-1", "discount-b"},
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
