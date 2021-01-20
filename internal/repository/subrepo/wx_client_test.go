package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/test"
	"github.com/objcoding/wxpay"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestWxPayClient_MockWebhookPayload(t *testing.T) {
	p := test.NewPersona().
		SetPayMethod(enum.PayMethodWx)
	order := p.NewOrder(enum.OrderKindCreate)
	account := p.FtcAccount()

	repo := test.NewRepo()
	repo.MustSaveAccount(account)
	repo.MustSaveOrder(order)
	t.Logf("Created order %s for user %s", order.ID, account.FtcID)

	client := NewWxPayClient(test.WxPayApp, zaptest.NewLogger(t))

	payload := client.MockWebhookPayload(test.NewWxWHUnsigned(order))

	t.Logf("Webhook raw payload: %s", wxpay.MapToXml(payload))
}
