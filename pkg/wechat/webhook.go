package wechat

import (
	"errors"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

// Notification is the data sent by wechat after payment
// finished.
// It is parsed from wechat's raw xml string.
type Notification struct {
	BaseResp
	OpenID        null.String `db:"open_id"`
	IsSubscribed  bool        `db:"is_subscribed"`
	TradeType     null.String `db:"trade_type"`
	BankType      null.String `db:"bank_type"`
	TotalFee      null.Int    `db:"total_fee"`
	Currency      null.String `db:"currency"`
	TransactionID null.String `db:"transaction_id"`
	FTCOrderID    null.String `db:"ftc_order_id"`
	TimeEnd       null.String `db:"time_end"`
}

// NewNotification converts wxpay.Params type to Notification type.
func NewNotification(p wxpay.Params) Notification {
	n := Notification{
		BaseResp: NewBaseResp(p),
	}

	v, ok := p["openid"]
	n.OpenID = null.NewString(v, ok)

	n.IsSubscribed = p.GetString("is_subscribe") == "Y"

	v, ok = p["trade_type"]
	n.TradeType = null.NewString(v, ok)

	v, ok = p["bank_type"]
	n.BankType = null.NewString(v, ok)

	fee := p.GetInt64("total_fee")
	n.TotalFee = null.NewInt(fee, fee != 0)

	v, ok = p["fee_type"]
	n.Currency = null.NewString(v, ok)

	v, ok = p["transaction_id"]
	n.TransactionID = null.NewString(v, ok)

	v, ok = p["out_trade_no"]
	n.FTCOrderID = null.NewString(v, ok)

	v, ok = p["time_end"]
	n.TimeEnd = null.NewString(v, ok)

	return n
}

func (n Notification) EnsureSuccess() error {
	if n.TotalFee.IsZero() {
		return errors.New("no payment amount found in wx webhook")
	}

	if n.FTCOrderID.IsZero() {
		return errors.New("no order id in wx webhook")
	}

	return nil
}
