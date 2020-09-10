package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"strings"
)

// PayInput defines the shared input fields for
// both alipay and wxpay.
type PayInput struct {
	// Tier and cycle are extracted from url.
	product.Edition
	// The following fields are not used yet.
	PlanID string `json:"planId"`
}

func (i *PayInput) Validate() *render.ValidationError {

	if i.Tier == enum.TierNull {
		return &render.ValidationError{
			Message: "Product edition is required",
			Field:   "tier",
			Code:    render.CodeMissingField,
		}
	}

	if i.Cycle == enum.CycleNull {
		return &render.ValidationError{
			Message: "Billing cycle is required",
			Field:   "cycle",
			Code:    render.CodeMissingField,
		}
	}

	return nil
}

type WxPayInput struct {
	PayInput
	// trade_type=JSAPI时（即JSAPI支付），此参数必传，此参数为微信用户在商户对应appid下的唯一标识。
	OpenID   null.String      `json:"openId"`
	Platform wechat.TradeType `json:"-"`
}

func NewWxPayInput(t wechat.TradeType) WxPayInput {
	return WxPayInput{
		Platform: t,
	}
}

func (i *WxPayInput) Validate() *render.ValidationError {
	if i.Platform == wechat.TradeTypeJSAPI && i.OpenID.IsZero() {
		return &render.ValidationError{
			Message: "Open id is required when calling JSAPI",
			Field:   "openId",
			Code:    render.CodeMissingField,
		}
	}

	return i.PayInput.Validate()
}

type AliPayInput struct {
	PayInput
	ReturnURL string `json:"returnUrl"` // Only required for desktop.
}

func (i *AliPayInput) Validate() *render.ValidationError {
	i.ReturnURL = strings.TrimSpace(i.ReturnURL)

	return i.PayInput.Validate()
}
