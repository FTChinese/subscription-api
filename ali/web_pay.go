package ali

import (
	"gitlab.com/ftchinese/subscription-api/paywall"
	"net/url"
)

type DesktopWebPay struct {
	FtcOrderID string  `json:"ftcOrderId"`
	ListPrice  float64 `json:"listPrice"`
	NetPrice   float64 `json:"netPrice"`
	// This is the url client should redirect to.
	// A simple `GET` works.
	PayURL string `json:"payUrl"`
}

func NewDesktopWebPay(s paywall.Subscription, payURL *url.URL) DesktopWebPay {
	return DesktopWebPay{
		FtcOrderID: s.OrderID,
		ListPrice:  s.ListPrice,
		NetPrice:   s.NetPrice,
		PayURL:     payURL.String(),
	}
}
