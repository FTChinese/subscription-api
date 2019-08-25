package test

import (
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"testing"
)

func TestMockWxXMLNotification(t *testing.T) {
	orderID, _ := paywall.GenerateOrderID()

	noti := WxXMLNotification(orderID)

	t.Logf("Mocked wxpay notifiction: %s", noti)
}

func TestWxNotification(t *testing.T) {
	orderID, _ := paywall.GenerateOrderID()

	noti := WxNotification(orderID)

	t.Logf("Notification: %+v", noti)
}

func TestWxXMLPrepay(t *testing.T) {
	prepay := WxXMLPrepay()

	t.Logf("Prepay response: %s", prepay)
}

func TestWxPrepay(t *testing.T) {
	uni := WxPrepay()

	t.Logf("Unified order: %+v", uni)
}

func TestGenCardSerial(t *testing.T) {
	t.Log(GenCardSerial())
}

func TestCreateGiftCard(t *testing.T) {

	m := NewRepo()

	tests := []struct {
		name string
	}{
		{
			name: "Create Gift Card",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.CreateGiftCard()

			t.Logf("Created gift card: %+v", got)
		})
	}
}
