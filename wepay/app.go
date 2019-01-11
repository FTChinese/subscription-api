package wepay

import (
	"github.com/objcoding/wxpay"
)

// PayApp is an app registered under Wechat.
type PayApp struct {
	AppID  string
	MchID  string
	APIKey string
	IsProd bool
}

// VerifyIdentity checks the identity of a Wechat response.
func (app PayApp) VerifyIdentity(params wxpay.Params) bool {
	if !params.ContainsKey("appid") || (params.GetString("appid") != app.AppID) {
		logger.WithField("trace", "VerifyIdentity").Error("Missing or wrong appid")
		return false
	}

	if !params.ContainsKey("mch_id") || (params.GetString("mch_id") != app.MchID) {
		logger.WithField("trace", "VerifyIdentity").Error("Missing or wrong mch_id")
		return false
	}

	return true
}
