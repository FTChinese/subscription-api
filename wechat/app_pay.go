package wechat

import (
	"github.com/objcoding/wxpay"
)

// AppPay is used by native app.
// Android <= 2.0.4 uses this structure.
// Actually we do not need to strictly conform to wechat's
// field naming convention, which is messy and arbitrary.
// `package` is a keyword in Java and Kotlin.
// Deprecate
type LegacyAppPay struct {
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

func (p LegacyAppPay) Param() wxpay.Params {
	param := make(wxpay.Params)
	param["appid"] = p.AppID
	param["partnerid"] = p.PartnerID
	param["prepayid"] = p.PrepayID
	param["package"] = p.Package
	param["noncestr"] = p.Nonce
	param["timestamp"] = p.Timestamp

	return param
}

func (p LegacyAppPay) WithHash(sign string) LegacyAppPay {
	p.Signature = sign
	return p
}

type AppPay struct {
	Pay
	PartnerID string `json:"partnerId"`
	PrepayID  string `json:"prepayId"`
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Package   string `json:"pkg"`
	Signature string `json:"signature"`
}

func (p AppPay) Params() wxpay.Params {
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
