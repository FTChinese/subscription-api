package repository

import "gitlab.com/ftchinese/subscription-api/models/wechat"

// SavePrepayResp saves Wechat prepay response for future analysis.
func (env Env) SavePrepayResp(orderID string, p wechat.UnifiedOrderResp) error {
	query := `
	INSERT INTO premium.log_wx_prepay
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
		prepay_id = ?,
	    code_url = ?,
	    mweb_url = ?,
	    created_utc = UTC_TIMESTAMP()`

	_, err := env.db.Exec(query,
		orderID,
		p.StatusCode,
		p.StatusMessage,
		p.AppID,
		p.MID,
		p.Nonce,
		p.Signature,
		p.ResultCode,
		p.ErrorCode,
		p.ErrorMessage,
		p.TradeType,
		p.PrepayID,
		p.CodeURL,
		p.MWebURL,
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

	_, err := env.db.Exec(query,
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

// SaveWxQueryResp stores wechat pay query result to DB.
func (env Env) SaveWxQueryResp(q wechat.OrderQueryResp) error {
	query := `
	INSERT IGNORE INTO premium.log_wx_order_query
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
		trade_type = ?,
	    trade_state = ?,
		bank_type = ?,
		total_fee = ?,
		currency = ?,
		transaction_id = ?,
		ftc_order_id = ?,
		time_end = ?,
	    trade_state_desc = ?`

	_, err := env.db.Exec(query,
		q.StatusCode,
		q.StatusMessage,
		q.AppID,
		q.MID,
		q.Nonce,
		q.Signature,
		q.ResultCode,
		q.ErrorCode,
		q.ErrorMessage,
		q.OpenID,
		q.TradeType,
		q.TradeState,
		q.BankType,
		q.TotalFee,
		q.Currency,
		q.TransactionID,
		q.FTCOrderID,
		q.TimeEnd,
		q.TradeStateDesc,
	)

	if err != nil {
		return err
	}

	return nil
}
