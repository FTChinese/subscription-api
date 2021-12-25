package subs

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
)

type PaymentIntent struct {
	Price      price.Price       `json:"price"`
	Offer      price.Discount    `json:"offer"`
	Order      Order             `json:"order"`
	Membership reader.Membership `json:"membership"`
}

type PaymentIntentSchema struct {
	OrderID      string                  `db:"order_id"`
	Price        price.JSONPrice         `db:"price"`
	Offer        price.ColumnDiscount    `db:"offer"`
	Membership   reader.ColumnMembership `db:"membership"`
	WxPayParams  wechat.ColumnSDKParams  `db:"wxpay_params"`
	AliPayParams ali.ColumnSDKParams     `db:"alipay_params"`
	CreatedUTC   chrono.Time             `db:"created_utc"`
}

type WxPaymentIntent struct {
	PaymentIntent
	Params wechat.SDKParams `json:"params" db:"wxpay_params"`
}

func NewWxPaymentIntent(pi PaymentIntent, params wechat.SDKParams) WxPaymentIntent {
	return WxPaymentIntent{
		PaymentIntent: pi,
		Params:        params,
	}
}

func (p WxPaymentIntent) Schema() PaymentIntentSchema {
	return PaymentIntentSchema{
		OrderID: p.Order.ID,
		Price: price.JSONPrice{
			Price: p.Price,
		},
		Offer: price.ColumnDiscount{
			Discount: p.Offer,
		},
		Membership: reader.ColumnMembership{
			Membership: p.Membership,
		},
		WxPayParams: wechat.ColumnSDKParams{
			SDKParams: p.Params,
		},
		AliPayParams: ali.ColumnSDKParams{},
		CreatedUTC:   chrono.TimeNow(),
	}
}

type AliPaymentIntent struct {
	PaymentIntent
	Params ali.SDKParams `json:"params" db:"alipay_params"`
}

func NewAliPaymentIntent(pi PaymentIntent, param string, kind ali.EntryKind) (AliPaymentIntent, error) {
	switch kind {
	case ali.EntryApp:
		return AliPaymentIntent{
			PaymentIntent: pi,
			Params: ali.SDKParams{
				BrowserRedirect: null.String{},
				AppSDK:          null.StringFrom(param),
			},
		}, nil

	case ali.EntryDesktopWeb, ali.EntryMobileWeb:
		return AliPaymentIntent{
			PaymentIntent: pi,
			Params: ali.SDKParams{
				BrowserRedirect: null.StringFrom(param),
				AppSDK:          null.String{},
			},
		}, nil
	}

	return AliPaymentIntent{}, errors.New("unknown alipay platform")
}

func (p AliPaymentIntent) Schema() PaymentIntentSchema {
	return PaymentIntentSchema{
		OrderID: p.Order.ID,
		Price: price.JSONPrice{
			Price: p.Price,
		},
		Offer: price.ColumnDiscount{
			Discount: p.Offer,
		},
		Membership: reader.ColumnMembership{
			Membership: p.Membership,
		},
		WxPayParams: wechat.ColumnSDKParams{},
		AliPayParams: ali.ColumnSDKParams{
			SDKParams: p.Params,
		},
		CreatedUTC: chrono.TimeNow(),
	}
}
