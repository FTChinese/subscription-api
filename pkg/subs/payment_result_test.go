package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/config"
	"testing"
)

func TestNewAliWebhookResult(t *testing.T) {
	config.MustSetupViper()

	order := MockOrder(faker.PlanStdYear, enum.OrderKindCreate)

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
