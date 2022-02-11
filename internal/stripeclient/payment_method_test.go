package stripeclient

import (
	"github.com/stripe/stripe-go/v72"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestClient_NewPaymentMethod(t *testing.T) {
	c := New(false, zaptest.NewLogger(t))

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
