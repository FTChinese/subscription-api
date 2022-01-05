//go:build !production
// +build !production

package test

import (
	"encoding/json"
	"encoding/xml"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/objcoding/wxpay"
	"strconv"
)

type WxWebhookPayload struct {
	XMLName    xml.Name `json:"-" xml:"xml"`
	ReturnCode string   `json:"return_code" xml:"return_code,omitempty"` // SUCCESS/FAIL
	ReturnMsg  string   `json:"return_msg,omitempty" xml:"return_msg,omitempty"`

	// 以下字段在return_code为SUCCESS的时候有返回
	AppID              string `json:"appid" xml:"appid"`
	MchID              string `json:"mch_id" xml:"mch_id"`
	DeviceInfo         string `json:"device_info,omitempty" xml:"device_info,omitempty"`
	Nonce              string `json:"nonce_str" xml:"nonce_str"`
	Sign               string `json:"sign" xml:"sign"`
	SignType           string `json:"sign_type" xml:"sign_type"`
	ResultCode         string `json:"result_code" xml:"result_code"`
	ErrCode            string `json:"err_code,omitempty" xml:"err_code,omitempty"`
	ErrCodeDes         string `json:"err_code_des,omitempty" xml:"err_code_des,omitempty"`
	OpenID             string `json:"openid" xml:"openid"`
	IsSubscribed       string `json:"is_subscribe" xml:"is_subscribe"` // 用户是否关注公众账号，Y-关注，N-未关注
	TradeType          string `json:"trade_type" xml:"trade_type"`
	BankType           string `json:"bank_type" xml:"bank_type"`
	TotalFee           string `json:"total_fee" xml:"total_fee"`
	SettlementTotalFee string `json:"settlement_total_fee,omitempty" xml:"settlement_total_fee,omitempty"`
	FeeType            string `json:"fee_type,omitempty" xml:"fee_type,omitempty"`
	CashFee            string `json:"cash_fee" xml:"cash_fee"`
	CashFeeType        string `json:"cash_fee_type,omitempty" xml:"cash_fee_type,omitempty"`
	CouponFee          string `json:"coupon_fee,omitempty" xml:"coupon_fee,omitempty"`
	CouponCount        string `json:"coupon_count,omitempty" xml:"coupon_count,omitempty"`
	CouponType         string `json:"coupon_type,omitempty" xml:"coupon_type,omitempty"`
	CouponID           string `json:"coupon_id_$n,omitempty" xml:"coupon_id,omitempty"`
	CouponFeeN         string `json:"coupon_fee_$n,omitempty" xml:"coupon_fee_$n,omitempty"`
	TransactionID      string `json:"transaction_id" xml:"transaction_id"`
	OutTradeNo         string `json:"out_trade_no,omitempty" xml:"out_trade_no,omitempty"`
	Attach             string `json:"attach,omitempty" xml:"attach,omitempty"`
	TimeEnd            string `json:"time_end" xml:"time_end"`
}

func NewWxWebhookPayload(o subs.Order) WxWebhookPayload {
	return WxWebhookPayload{
		ReturnCode:         wechat.Success,
		ReturnMsg:          "",
		AppID:              WxPayApp.AppID,
		MchID:              WxPayApp.MchID,
		DeviceInfo:         "",
		Nonce:              wechat.GenerateNonce(),
		Sign:               "",
		SignType:           wechat.SignTypeMD5,
		ResultCode:         wechat.Success,
		ErrCode:            "",
		ErrCodeDes:         "",
		OpenID:             faker.GenWxID(),
		IsSubscribed:       "N",
		TradeType:          string(wechat.TradeTypeApp),
		BankType:           "ICBC_DEBIT",
		TotalFee:           strconv.FormatInt(o.WxPayable(), 10),
		SettlementTotalFee: "",
		FeeType:            "",
		CashFee:            strconv.FormatInt(o.WxPayable(), 10),
		CashFeeType:        "",
		CouponFee:          "",
		CouponCount:        "",
		CouponType:         "",
		CouponID:           "",
		CouponFeeN:         "",
		TransactionID:      rand.StringWithCharset(32, "1234567890"),
		OutTradeNo:         o.ID,
		Attach:             "",
		TimeEnd:            wechat.GenerateTimestamp(),
	}
}

func (p WxWebhookPayload) WithSign() WxWebhookPayload {
	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	params := make(wxpay.Params)
	err = json.Unmarshal(b, &params)
	if err != nil {
		panic(err)
	}

	p.Sign = WxPayClient.Sign(params)

	return p
}

func (p WxWebhookPayload) ToMap() wxpay.Params {
	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	params := make(wxpay.Params)
	err = json.Unmarshal(b, &params)
	if err != nil {
		panic(err)
	}

	params["sign"] = WxPayClient.Sign(params)

	return params
}

func (p WxWebhookPayload) ToXML() string {
	b, err := xml.MarshalIndent(p.WithSign(), "", "\t")

	if err != nil {
		panic(err)
	}

	return string(b)
}
