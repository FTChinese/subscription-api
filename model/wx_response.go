package model

import "gitlab.com/ftchinese/subscription-api/wechat"

// SavePrepayResp saves Wechat prepay response for future analysis.
func (env Env) SavePrepayResp(orderID string, p wechat.UnifiedOrderResp) error {
	query := `
	INSERT IGNORE INTO premium.log_wx_prepay
	SET order_id = ?,
		return_code = ?,
		return_message = ?,
		app_id = ?,
		merchant_id = ?,
		nonce = ?,
		signature = ?,
		result_code = ?,
		error_code = ?,
		error_description = ?,
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
		p.ResultCode,
		p.ErrorCode,
		p.ErrorDescription,
		p.TradeType,
		p.PrepayID,
	)

	if err != nil {
		return err
	}

	return nil
}

// SaveWxNotification saves a wechat notification for logging purpose.
func (env Env) SaveWxNotification(n wechat.Notification) error {
	query := `
	INSERT IGNORE INTO premium.log_wx_notification
	SET return_code = ?,
		return_message = ?,
		app_id = ?,
		merchant_id = ?,
		nonce = ?,
		signature = ?,
		result_code = ?,
		error_code = ?,
		error_description = ?,
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
		n.ResultCode,
		n.ErrorCode,
		n.ErrorMessage,
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
