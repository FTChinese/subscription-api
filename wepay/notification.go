package wepay

import (
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

// Notification contains wechat's notification data after payment finished.
type Notification struct {
	StatusCode    string
	StatusMessage string
	AppID         null.String
	MID           null.String
	Nonce         null.String
	Signature     null.String
	IsSuccess     bool
	ResultCode    null.String
	ResultMessage null.String
	OpenID        null.String
	IsSubscribed  bool
	TradeType     null.String
	BankType      null.String
	TotalFee      null.Int
	Currency      null.String
	TransactionID null.String
	FTCOrderID    null.String
	CreatedAt     null.String
}

// NewNotification converts wxpay.Params type to Notification type.
func NewNotification(r wxpay.Params) Notification {
	n := Notification{
		StatusCode:    r.GetString("return_code"),
		StatusMessage: r.GetString("return_msg"),
		IsSuccess:     r.GetString("result_code") == "SUCCESS",
		IsSubscribed:  r.GetString("is_subscribe") == "Y",
	}

	if v, ok := r["appid"]; ok {
		n.AppID = null.StringFrom(v)
	}

	if v, ok := r["mch_id"]; ok {
		n.MID = null.StringFrom(v)
	}

	if v, ok := r["nonce_str"]; ok {
		n.Nonce = null.StringFrom(v)
	}

	if v, ok := r["sign"]; ok {
		n.Signature = null.StringFrom(v)
	}

	if v, ok := r["err_code"]; ok {
		n.ResultCode = null.StringFrom(v)
	}
	if v, ok := r["err_code_des"]; ok {
		n.ResultMessage = null.StringFrom(v)
	}

	if v, ok := r["trade_type"]; ok {
		n.TradeType = null.StringFrom(v)
	}

	if v, ok := r["bank_type"]; ok {
		n.BankType = null.StringFrom(v)
	}

	if v := r.GetInt64("total_fee"); v != 0 {
		n.TotalFee = null.IntFrom(v)
	}

	if v, ok := r["fee_type"]; ok {
		n.Currency = null.StringFrom(v)
	}

	if v, ok := r["transaction_id"]; ok {
		n.TransactionID = null.StringFrom(v)
	}

	if v, ok := r["out_trade_no"]; ok {
		n.FTCOrderID = null.StringFrom(v)
	}

	if v, ok := r["time_end"]; ok {
		n.CreatedAt = null.StringFrom(v)
	}

	return n
}

// SaveNotification saves a wechat notification for logging purpose.
func (env Env) SaveNotification(n Notification) error {
	query := `
	INSERT INTO premium.log_wx_notification
	SET status_code = ?,
		status_message = ?,
		app_id = ?,
		merchant_id = ?,
		nonce = ?,
		signature = ?,
		is_success = ?,
		result_code = ?,
		result_message = ?,
		open_id = ?,
		is_subscribed = ?,
		trade_type = ?,
		bank_type = ?,
		total_fee = ?,
		currency = ?,
		transaction_id = ?,
		ftc_order_id = ?,
		created_utc = ?`

	_, err := env.DB.Exec(query,
		n.StatusCode,
		n.StatusMessage,
		n.AppID,
		n.MID,
		n.Nonce,
		n.Signature,
		n.IsSuccess,
		n.ResultCode,
		n.ResultMessage,
		n.OpenID,
		n.IsSubscribed,
		n.TradeType,
		n.BankType,
		n.TotalFee,
		n.Currency,
		n.TransactionID,
		n.FTCOrderID,
		n.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}
