package model

import "gitlab.com/ftchinese/subscription-api/wechat"

// SavePrepayResp saves Wechat prepay response for future analysis.
func (env Env) SavePrepayResp(orderID string, p wechat.PrepayResp) error {
	query := `
	INSERT INTO premium.log_wx_prepay
	SET order_id = ?,
		status_code = ?,
		status_message = ?,
		app_id = ?,
		merchant_id = ?,
		nonce = ?,
		signature = ?,
		is_success = ?,
		result_code = ?,
		result_message = ?,
		trade_type = ?,
		prepay_id = ?`

	_, err := env.DB.Exec(query,
		orderID,
		p.StatusCode,
		p.StatusMessage,
		p.AppID,
		p.MID,
		p.Nonce,
		p.Signature,
		p.IsSuccess,
		p.ResultCode,
		p.ResultMessage,
		p.TradeType,
		p.PrePayID,
	)

	if err != nil {
		return err
	}

	return nil
}

// SaveWxNotification saves a wechat notification for logging purpose.
func (env Env) SaveWxNotification(n wechat.Notification) error {
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
		time_end = ?`

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
		n.TimeEnd,
	)

	if err != nil {
		return err
	}

	return nil
}
