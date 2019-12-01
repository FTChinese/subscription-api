package test

import (
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"testing"
)

func TestMockWxXMLNotification(t *testing.T) {
	orderID, _ := subscription.GenerateOrderID()

	noti := WxXMLNotification(orderID)

	t.Logf("Mocked wxpay notifiction: %s", noti)
}

func TestWxNotification(t *testing.T) {
	orderID, _ := subscription.GenerateOrderID()

	noti := WxNotification(orderID)

	t.Logf("Notification: %+v", noti)
}

func TestWxXMLPrepay(t *testing.T) {
	prepay := WxXMLPrepay()

	t.Logf("Prepay response: %s", prepay)
}

func TestWxPrepay(t *testing.T) {
	uni := WxPrepay(MustGenOrderID())

	t.Logf("Unified order: %+v", uni)
}

func TestGenCardSerial(t *testing.T) {
	t.Log(GenCardSerial())
}
