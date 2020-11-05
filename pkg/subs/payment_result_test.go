package subs

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"testing"
)

func TestNewAliWebhookResult(t *testing.T) {
	config.MustSetupViper()

	order := MockOrder()

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
