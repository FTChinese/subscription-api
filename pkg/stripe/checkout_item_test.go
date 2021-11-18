package stripe

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/guregu/null"
	"testing"
)

func TestCheckoutItem_NewSubParams(t *testing.T) {
	item := CheckoutItem{
		Price:        MockPriceStdYear,
		Introductory: MockPriceStdIntro,
	}

	p := item.NewSubParams("customer-id", SubSharedParams{
		CouponID:             null.String{},
		DefaultPaymentMethod: null.String{},
		IdempotencyKey:       "key",
	})

	t.Logf("%s", faker.MustMarshalIndent(p))
}
