package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/test"
	"github.com/objcoding/wxpay"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestWxPayClient_MockWebhookPayload(t *testing.T) {
	p := test.NewPersona()
	order := p.OrderBuilder().
		WithPayMethod(enum.PayMethodWx).
		WithKind(enum.OrderKindCreate).
		Build()

	account := p.EmailOnlyAccount()

	repo := test.NewRepo()
	repo.MustCreateFtcAccount(account)
	repo.MustSaveOrder(order)
	t.Logf("Created order %s for user %s", order.ID, account.FtcID)

	client := NewWxPayClient(test.WxPayApp, zaptest.NewLogger(t))

	payload := client.MockWebhookPayload(test.NewWxWHUnsigned(order))

	t.Logf("Webhook raw payload: %s", wxpay.MapToXml(payload))
}
