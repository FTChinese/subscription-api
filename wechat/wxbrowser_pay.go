package wechat

import "github.com/objcoding/wxpay"

// WxBrowserPay is used inside wechat browser
type WxBrowserPay struct {
	FtcOrderID string  `json:"ftcOrderId"`
	ListPrice  float64 `json:"listPrice"`
	NetPrice   float64 `json:"netPrice"`
	AppID      string  `json:"appId"`
	Timestamp  string  `json:"timestamp"`
	Nonce      string  `json:"nonce"`
	Package    string  `json:"package"`
	SignType   string  `json:"signType"`
	Signature  string  `json:"signature"`
}

func (p WxBrowserPay) Params() wxpay.Params {
	param := make(wxpay.Params)
	param["appid"] = p.AppID
	param["timeStamp"] = p.Timestamp
	param["nonceStr"] = p.Nonce
	param["package"] = p.Package
	param["signType"] = p.SignType

	return param
}

func (p WxBrowserPay) WithHash(sign string) WxBrowserPay {
	p.Signature = sign
	return p
}
