package subrepo

import "fmt"

const (
	InsertAliPayLoad = `
	INSERT INTO premium.log_ali_notification
	SET notified_cst = ?,
		notify_type = ?,
		notify_id = ?,
		app_id = ?,
		charset = ?,
		version = ?,
		sign_type = ?,
		sign = ?,
		trade_number = ?,
		ftc_order_id = ?,
		out_biz_no = NULLIF(?, ''),
		buyer_id = NULLIF(?, ''),
		buyer_login_id = NULLIF(?, ''),
		seller_id = NULLIF(?, ''),
		seller_email = NULLIF(?, ''),
		trade_status = NULLIF(?, ''),
		total_amount = NULLIF(?, ''),
		receipt_amount = NULLIF(?, ''),
		invoice_amount = NULLIF(?, ''),
		buyer_pay_amount = NULLIF(?, ''),
		point_amount = NULLIF(?, ''),
		refund_fee = NULLIF(?, ''),
		created_cst = NULLIF(?, ''),
		paid_cst = NULLIF(?, ''),
		refunded_cst = NULLIF(?, ''),
		closed_cst = NULLIF(?, ''),
		fund_bill_list = NULLIF(?, ''),
		passback_param = NULLIF(?, ''),
		voucher_detail_list = NULLIF(?, '')`

	stmtWxBaseResp = `
	return_code = :status_code,
	return_message = :status_message,
	app_id = :app_id,
	merchant_id = :merchant_id,
	nonce = :nonce,
	signature = :signature,
	result_code = :result_code,
	error_code = :error_code,
	error_description = :error_message`
)

var (
	InsertWxPrepay = fmt.Sprintf(`
	INSERT INTO premium.log_wx_prepay
	SET %s,
		trade_type = :trade_type,
		prepay_id = :prepay_id,
	    code_url = :qr_code,
	    mweb_url = :mobile_redirect_url,
		order_id = :order_id,
	    created_utc = UTC_TIMESTAMP()`, stmtWxBaseResp)

	InsertWxPayLoad = fmt.Sprintf(`
	INSERT IGNORE INTO premium.log_wx_notification
	SET %s,
		open_id = :open_id,
		is_subscribed = :is_subscribed,
		trade_type = :trade_type,
		bank_type = :bank_type,
		total_fee = :total_fee,
		currency = :currency,
		transaction_id = :transaction_id,
		ftc_order_id = :ftc_order_id,
		time_end = :time_end`, stmtWxBaseResp)

	InsertWxQueryPayLoad = fmt.Sprintf(`
	INSERT IGNORE INTO premium.log_wx_order_query
	SET %s,
		open_id = :open_id,
		trade_type = :trade_type,
	    trade_state = :trade_state,
		trade_state_desc = :trade_state_desc,
		bank_type = :bank_type,
		total_fee = :total_fee,
		currency = :currency,
		transaction_id = :transaction_id,
		ftc_order_id = :ftc_order_id,
		time_end = :time_end`, stmtWxBaseResp)
)
