package ali

import (
	"gitlab.com/ftchinese/subscription-api/paywall"
	"net/url"
)

type WebPay struct {
	FtcOrderID string  `json:"ftcOrderId"`
	ListPrice  float64 `json:"listPrice"`
	NetPrice   float64 `json:"netPrice"`
	// This is the url client should redirect to.
	// A simple `GET` works.
	PayURL string `json:"payUrl"`
}

func NewWebPay(s paywall.Subscription, payURL *url.URL) WebPay {
	return WebPay{
		FtcOrderID: s.OrderID,
		ListPrice:  s.ListPrice,
		NetPrice:   s.NetPrice,
		PayURL:     payURL.String(),
	}
}
