package wechat

import (
	"errors"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"time"
)

// Notification is the data sent by wechat after payment
// finished.
// It is parsed from wechat's raw xml string.
type Notification struct {
	Resp
	OpenID        null.String `db:"open_id"`
	IsSubscribed  bool        `db:"is_subscribed"`
	TradeType     null.String `db:"trade_type"`
	BankType      null.String `db:"bank_type"`
	TotalFee      null.Int    `db:"total_fee"`
	Currency      null.String `db:"currency"`
	TransactionID null.String `db:"transaction_id"`
	FTCOrderID    null.String `db:"ftc_order_id"`
	TimeEnd       null.String `db:"time_end"`
	params        wxpay.Params
}

// Params turns the struct into wxpay.Param so that we
// could generate a signature.
// This is used for mocking only.
func (n Notification) Params() wxpay.Params {
	p := n.BaseParams()

	var subscribed string
	if n.IsSubscribed {
		subscribed = "Y"
	} else {
		subscribed = "N"
	}

	p.SetString("openid", n.OpenID.String)
	p.SetString("is_subscribe", subscribed)
	p.SetString("bank_type", n.BankType.String)
	p.SetInt64("total_fee", n.TotalFee.Int64)
	p.SetString("transaction_id", n.TransactionID.String)
	p.SetString("out_trade_no", n.FTCOrderID.String)
	p.SetString("time_end", n.TimeEnd.String)

	return p
}

// NewNotification converts wxpay.Params type to Notification type.
func NewNotification(p wxpay.Params) Notification {
	n := Notification{}

	n.Populate(p)

	if v, ok := p["openid"]; ok {
		n.OpenID = null.StringFrom(v)
	}

	n.IsSubscribed = p.GetString("is_subscribe") == "Y"

	if v, ok := p["trade_type"]; ok {
		n.TradeType = null.StringFrom(v)
	}

	if v, ok := p["bank_type"]; ok {
		n.BankType = null.StringFrom(v)
	}

	if v := p.GetInt64("total_fee"); v != 0 {
		n.TotalFee = null.IntFrom(v)
	}

	if v, ok := p["fee_type"]; ok {
		n.Currency = null.StringFrom(v)
	}

	if v, ok := p["transaction_id"]; ok {
		n.TransactionID = null.StringFrom(v)
	}

	if v, ok := p["out_trade_no"]; ok {
		n.FTCOrderID = null.StringFrom(v)
	}

	if v, ok := p["time_end"]; ok {
		n.TimeEnd = null.StringFrom(v)
	}

	n.params = p

	return n
}

func (n Notification) IsPriceMatched(cent int64) bool {
	if n.TotalFee.IsZero() {
		return false
	}

	return n.TotalFee.Int64 == cent
}

func (n Notification) GetPaymentResult() (subscription.PaymentResult, error) {
	if n.TotalFee.IsZero() {
		return subscription.PaymentResult{}, errors.New("no payment amount found in wx webhook")
	}

	if n.FTCOrderID.IsZero() {
		return subscription.PaymentResult{}, errors.New("no order id in wx webhook")
	}

	confirmedAt, err := util.ParseWxTime(n.TimeEnd.String)
	if err != nil {
		confirmedAt = time.Now()
	}

	return subscription.PaymentResult{
		Amount:      n.TotalFee.Int64,
		OrderID:     n.FTCOrderID.String,
		ConfirmedAt: confirmedAt,
	}, nil
}
