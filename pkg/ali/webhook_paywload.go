package ali

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/smartwalle/alipay"
)

type WebhookPayload struct {
	OrderID    string
	Payload    ColumnPayload
	CreatedUTC chrono.Time
}

func NewWebhookPayload(n *alipay.TradeNotification) WebhookPayload {
	return WebhookPayload{
		OrderID: n.OutTradeNo,
		Payload: ColumnPayload{
			TradeNotification: n,
		},
		CreatedUTC: chrono.TimeNow(),
	}
}

const StmtSavePayload = `
INSERT INTO premium.alipay_webhook
SET order_id = :order_id,
	payload = :payload,
	created_utc = :created_utc
`

// StmtInsertAliPayLoad saves alipay notification
const StmtInsertAliPayLoad = `
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
	voucher_detail_list = NULLIF(?, '')
`
