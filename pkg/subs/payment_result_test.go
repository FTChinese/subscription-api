package subs

import (
	"github.com/FTChinese/subscription-api/faker"
	"testing"
)

func TestNewAliWebhookResult(t *testing.T) {
	faker.MustSetupViper()

	order := NewMockOrderBuilder("").Build()

	n := MockAliNoti(order)
	pr, err := NewAliWebhookResult(&n)

	if err != nil {
		t.Error(err)
		return
	}

	err = order.ValidatePayment(pr)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%v", pr)
}
