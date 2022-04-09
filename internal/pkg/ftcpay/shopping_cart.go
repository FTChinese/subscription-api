package ftcpay

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"strings"
)

type WxPayReq struct {
	FtcCartParams
	// trade_type=JSAPI时（即JSAPI支付），此参数必传，此参数为微信用户在商户对应appid下的唯一标识。
	OpenID   null.String      `json:"openId"`
	Platform wechat.TradeType `json:"-"`
}

func NewWxPayReq(t wechat.TradeType) WxPayReq {
	return WxPayReq{
		FtcCartParams: FtcCartParams{},
		OpenID:        null.String{},
		Platform:      t,
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

	return r.FtcCartParams.Validate()
}

type AliPayReq struct {
	FtcCartParams
	ReturnURL string `json:"returnUrl"` // Only required for desktop.
}

func (r AliPayReq) Validate() *render.ValidationError {
	r.ReturnURL = strings.TrimSpace(r.ReturnURL)

	return r.FtcCartParams.Validate()
}
