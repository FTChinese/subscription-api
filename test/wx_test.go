package test

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"testing"
)

func TestWxXMLNotification(t *testing.T) {
	p := NewPersona().SetPayMethod(enum.PayMethodWx)

	repo := NewRepo()

	order := p.CreateOrder()
	account := p.Account()

	repo.MustSaveAccount(account)
	repo.MustSaveOrder(order)

	payload := WxXMLNotification(order)

	t.Logf("Created order %s for user %s", order.ID, account.FtcID)

	t.Logf("Wx webhook payload %s", payload)
}

func TestWxXMLPrepay(t *testing.T) {
	prepay := WxXMLPrepay()

	t.Logf("Prepay response: %s", prepay)
}

func TestWxPrepay(t *testing.T) {
	uni := WxPrepay(subs.MustGenerateOrderID())

	t.Logf("Unified order: %+v", uni)
}
