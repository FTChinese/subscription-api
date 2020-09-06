package wechat

import (
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

type OrderQueryResp struct {
	Resp
	OpenID    null.String `db:"open_id"`
	TradeType null.String `db:"trade_type"` // APP
	// SUCCESS
	// REFUND
	// NOTPAY
	// CLOSED
	// REVOKED
	// USERPAYING
	// PAYERROR
	// This is `trade_state` field in wechat response.
	TradeState null.String `db:"trade_state"`
	// This is the `trade_state_desc` field.
	TradeStateDesc null.String `db:"trade_state_desc"`
	BankType       null.String `db:"bank_type"`
	TotalFee       null.Int    `db:"total_fee"`
	Currency       null.String `db:"currency"`
	TransactionID  null.String `db:"transaction_id"`
	FTCOrderID     null.String `db:"ftc_order_id"`
	TimeEnd        null.String `db:"time_end"` // 20141030133525
}

func NewOrderQueryResp(p wxpay.Params) OrderQueryResp {
	r := OrderQueryResp{}

	r.Populate(p)

	if v, ok := p["openid"]; ok {
		r.OpenID = null.StringFrom(v)
	}

	if v, ok := p["trade_type"]; ok {
		r.TradeType = null.StringFrom(v)
	}

	if v, ok := p["trade_state"]; ok {
		r.TradeState = null.StringFrom(v)
	}

	if v, ok := p["bank_type"]; ok {
		r.BankType = null.StringFrom(v)
	}

	if price := p.GetInt64("total_fee"); price != 0 {
		r.TotalFee = null.IntFrom(price)
	}

	if v, ok := p["fee_type"]; ok {
		r.Currency = null.StringFrom(v)
	}

	if v, ok := p["transaction_id"]; ok {
		r.TransactionID = null.StringFrom(v)
	}

	if v, ok := p["out_trade_no"]; ok {
		r.FTCOrderID = null.StringFrom(v)
	}

	if v, ok := p["time_end"]; ok {
		r.TimeEnd = null.StringFrom(v)
	}

	if v, ok := p["trade_state_desc"]; ok {
		r.TradeStateDesc = null.StringFrom(v)
	}

	return r
}
