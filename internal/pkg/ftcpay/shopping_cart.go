package ftcpay

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"strings"
)

// OrderParams contains the item user want to buy.
// Both price and offer only requires id field to be set.
type OrderParams struct {
	PriceID    string         `json:"priceId"`
	DiscountID null.String    `json:"discountId"`
	Price      price.Price    `json:"price"`
	Offer      price.Discount `json:"offer"` // Optional
}

func (s *OrderParams) Validate() *render.ValidationError {
	if s.PriceID == "" {
		return &render.ValidationError{
			Message: "Missing priceId field",
			Field:   "priceId",
			Code:    render.CodeMissingField,
		}
	}

	return nil
}

type WxPayReq struct {
	OrderParams
	// trade_type=JSAPI时（即JSAPI支付），此参数必传，此参数为微信用户在商户对应appid下的唯一标识。
	OpenID   null.String      `json:"openId"`
	Platform wechat.TradeType `json:"-"`
}

func NewWxPayReq(t wechat.TradeType) WxPayReq {
	return WxPayReq{
		OrderParams: OrderParams{},
		OpenID:      null.String{},
		Platform:    t,
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

	return r.OrderParams.Validate()
}

type AliPayReq struct {
	OrderParams
	ReturnURL string `json:"returnUrl"` // Only required for desktop.
}

func (r AliPayReq) Validate() *render.ValidationError {
	r.ReturnURL = strings.TrimSpace(r.ReturnURL)

	return r.OrderParams.Validate()
}
