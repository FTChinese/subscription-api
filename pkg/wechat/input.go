package wechat

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/guregu/null"
)

// PayInput defines the request data when creating a wxpay order.
type PayInput struct {
	// trade_type=JSAPI时（即JSAPI支付），此参数必传，此参数为微信用户在商户对应appid下的唯一标识。
	OpenID   null.String `json:"openId"`
	Platform TradeType   `json:"-"`
}

func NewPayInput(t TradeType) PayInput {
	return PayInput{
		OpenID:   null.String{},
		Platform: t,
	}
}

func (i PayInput) Validate() *render.ValidationError {
	if i.Platform == TradeTypeJSAPI && i.OpenID.IsZero() {
		return &render.ValidationError{
			Message: "Open id is required when calling JSAPI",
			Field:   "openId",
			Code:    render.CodeMissingField,
		}
	}

	return nil
}
