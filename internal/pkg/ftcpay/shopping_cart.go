package ftcpay

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"strings"
)

// ShoppingCart contains the item user want to buy.
// Both price and offer only requires id field to be set.
type ShoppingCart struct {
	Price price.Price    `json:"price"`
	Offer price.Discount `json:"offer"` // Optional
}

func (s *ShoppingCart) Validate() *render.ValidationError {
	if s.Price.ID == "" {
		return &render.ValidationError{
			Message: "ID of price-to-subscribe not provided",
			Field:   "price.id",
			Code:    render.CodeMissingField,
		}
	}

	return nil
}

type WxPayReq struct {
	ShoppingCart
	// trade_type=JSAPI时（即JSAPI支付），此参数必传，此参数为微信用户在商户对应appid下的唯一标识。
	OpenID   null.String      `json:"openId"`
	Platform wechat.TradeType `json:"-"`
}

func NewWxPayReq(t wechat.TradeType) WxPayReq {
	return WxPayReq{
		ShoppingCart: ShoppingCart{},
		OpenID:       null.String{},
		Platform:     t,
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

	return r.ShoppingCart.Validate()
}

type AliPayReq struct {
	ShoppingCart
	ReturnURL string `json:"returnUrl"` // Only required for desktop.
}

func (r AliPayReq) Validate() *render.ValidationError {
	r.ReturnURL = strings.TrimSpace(r.ReturnURL)

	return r.ShoppingCart.Validate()
}
