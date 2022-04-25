package stripe

import (
	"testing"

	"github.com/FTChinese/subscription-api/faker"
)

func TestPaymentIntentJSON(t *testing.T) {
	pi := PaymentIntent{}

	b := faker.MustMarshalIndent(pi)

	t.Logf("%s", b)
}
