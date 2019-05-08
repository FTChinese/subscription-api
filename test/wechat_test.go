package test

import (
	"gitlab.com/ftchinese/subscription-api/paywall"
	"testing"
)

func TestMockWxXMLNotification(t *testing.T) {
	orderID, _ := paywall.GenerateOrderID()

	noti := MockWxXMLNotification(orderID)

	t.Logf("Mocked wxpay notifiction: %s", noti)
}

func TestWxNotification(t *testing.T) {
	orderID, _ := paywall.GenerateOrderID()

	noti := MockWxNotification(orderID)

	t.Logf("Notification: %+v", noti)
}

func TestWxXMLPrepay(t *testing.T) {
	prepay := MockWxXMLPrepay()

	t.Logf("Prepay response: %s", prepay)
}

func TestWxPrepay(t *testing.T) {
	uni := MockWxPrepay()

	t.Logf("Unified order: %+v", uni)
}
