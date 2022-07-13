//go:build !production

package test

import (
	"time"

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
