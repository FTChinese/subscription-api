package wechat

import (
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

// Notification contains wechat's notification data after payment finished.
type Notification struct {
	WxResp
	OpenID        null.String
	IsSubscribed  bool
	TradeType     null.String
	BankType      null.String
	TotalFee      null.Int
	Currency      null.String
	TransactionID null.String
	FTCOrderID    null.String
	TimeEnd       null.String
}

// NewNotification converts wxpay.Params type to Notification type.
func NewNotification(p wxpay.Params) Notification {
	n := Notification{

	}

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

	return n
}
