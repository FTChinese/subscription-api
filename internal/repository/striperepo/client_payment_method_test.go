package striperepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/stripe/stripe-go/v72"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestClient_NewPaymentMethod(t *testing.T) {
	c := NewClient(false, zaptest.NewLogger(t))

	pm, err := c.NewPaymentMethod(&stripe.PaymentMethodCardParams{
		CVC:      stripe.String("001"),
		ExpMonth: stripe.String("02"),
		ExpYear:  stripe.String("22"),
		Number:   stripe.String("4242424242424242"),
		Token:    nil,
	})

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", pm)
}

func TestClient_AttachPaymentMethod(t *testing.T) {
	faker.SeedGoFake()

	client := NewClient(false, zaptest.NewLogger(t))

	cus, err := client.CreateCustomer(gofakeit.Email())
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Customer email %s", cus.Email)

	pm, err := client.NewPaymentMethod(&stripe.PaymentMethodCardParams{
		CVC:      stripe.String("001"),
		ExpMonth: stripe.String("02"),
		ExpYear:  stripe.String("22"),
		Number:   stripe.String("4242424242424242"),
		Token:    nil,
	})

	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Payment method: %+v", pm)

	si, err := client.AttachPaymentMethod(cus.ID, pm.ID)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Setup intent: %+v", si)
}
