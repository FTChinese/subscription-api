package wechat

import "github.com/objcoding/wxpay"

// WxBrowserPay is used inside wechat browser
type WxBrowserPay struct {
	Pay
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Package   string `json:"pkg"`
	Signature string `json:"signature"`
	SignType  string `json:"signType"`
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
