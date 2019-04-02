package wechat

import (
	"github.com/objcoding/wxpay"
)

// AppPay is used by native app.
type AppPay struct {
	FtcOrderID string  `json:"ftcOrderId"`
	Price      float64 `json:"price"`
	ListPrice  float64 `json:"listPrice"`
	NetPrice   float64 `json:"netPrice"`
	AppID      string  `json:"appid"`
	PartnerID  string  `json:"partnerid"`
	PrepayID   string  `json:"prepayid"`
	Package    string  `json:"package"`
	Nonce      string  `json:"noncestr"`
	Timestamp  string  `json:"timestamp"`
	Signature  string  `json:"sign"`
}

func (p AppPay) Param() wxpay.Params {
	param := make(wxpay.Params)
	param["appid"] = p.AppID
	param["partnerid"] = p.PartnerID
	param["prepayid"] = p.PrepayID
	param["package"] = p.Package
	param["noncestr"] = p.Nonce
	param["timestamp"] = p.Timestamp

	return param
}

func (p AppPay) WithHash(sign string) AppPay {
	p.Signature = sign
	return p
}
