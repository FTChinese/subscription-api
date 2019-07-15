package controller

import (
	"github.com/FTChinese/go-rest/view"
	"gitlab.com/ftchinese/subscription-api/wechat"
)

type wxpayBody struct {
	PlanID string `json:"planId"`
	OpenID string `json:"openId"`
}

func (w wxpayBody) validate(tradeType wechat.TradeType) *view.Reason {
	if w.PlanID == "" {
		r := view.NewReason()
		r.Field = "planId"
		r.Code = view.CodeMissingField
		r.SetMessage("Please select a plan to subscribe")
		return r
	}

	if tradeType == wechat.TradeTypeJSAPI && w.OpenID == "" {
		r := view.NewReason()
		r.Field = "openId"
		r.Code = view.CodeMissingField
		r.SetMessage("You must provide open id to use wechat js api")

		return r
	}
	return nil
}
