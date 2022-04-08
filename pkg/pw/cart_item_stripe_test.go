package pw

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/guregu/null"
	"testing"
)

func TestCheckoutItem_NewSubParams(t *testing.T) {
	item := CartItemStripe{
		Recurring:    stripe.MockPriceStdYear,
		Introductory: stripe.MockPriceStdIntro,
	}

	p := item.NewSubParams("customer-id", StripeSubsParams{
		CouponID:             null.String{},
		DefaultPaymentMethod: null.String{},
		IdempotencyKey:       "key",
	})

	t.Logf("%s", faker.MustMarshalIndent(p))
}
