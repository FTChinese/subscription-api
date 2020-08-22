package test

import (
	"github.com/FTChinese/subscription-api/pkg/subs"
	"testing"
)

func TestMockWxXMLNotification(t *testing.T) {
	orderID, _ := subs.GenerateOrderID()

	noti := WxXMLNotification(orderID)

	t.Logf("Mocked wxpay notifiction: %s", noti)
}

func TestWxNotification(t *testing.T) {
	orderID, _ := subs.GenerateOrderID()

	noti := WxNotification(orderID)

	t.Logf("Notification: %+v", noti)
}

func TestWxXMLPrepay(t *testing.T) {
	prepay := WxXMLPrepay()

	t.Logf("Prepay response: %s", prepay)
}

func TestWxPrepay(t *testing.T) {
	uni := WxPrepay(subs.MustGenerateOrderID())

	t.Logf("Unified order: %+v", uni)
}
