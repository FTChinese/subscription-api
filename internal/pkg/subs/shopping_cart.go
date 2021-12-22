package subs

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"strings"
)

type WxPayReq struct {
	pw.CartParams
	// trade_type=JSAPI时（即JSAPI支付），此参数必传，此参数为微信用户在商户对应appid下的唯一标识。
	OpenID   null.String      `json:"openId"`
	Platform wechat.TradeType `json:"-"`
}

func NewWxPayReq(t wechat.TradeType) WxPayReq {
	return WxPayReq{
		CartParams: pw.CartParams{},
		OpenID:     null.String{},
		Platform:   t,
	}
}

func (r *WxPayReq) Validate() *render.ValidationError {
	if r.Platform == wechat.TradeTypeJSAPI && r.OpenID.IsZero() {
		return &render.ValidationError{
			Message: "Open id is required when calling JSAPI",
			Field:   "openId",
			Code:    render.CodeMissingField,
		}
	}

	return r.CartParams.Validate()
}

type AliPayReq struct {
	pw.CartParams
	ReturnURL string `json:"returnUrl"` // Only required for desktop.
}

func (r AliPayReq) Validate() *render.ValidationError {
	r.ReturnURL = strings.TrimSpace(r.ReturnURL)

	return r.CartParams.Validate()
}
