package model

import (
	"github.com/smartwalle/alipay"
)

// SaveAliNotification logs everything Alipay sends.
func (env Env) SaveAliNotification(n alipay.TradeNotification) error {
	query := `
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

	_, err := env.db.Exec(query,
		n.NotifyTime,
		n.NotifyType,
		n.NotifyId,
		n.AppId,
		n.Charset,
		n.Version,
		n.SignType,
		n.Sign,
		n.TradeNo,
		n.OutTradeNo,
		n.OutBizNo,
		n.BuyerId,
		n.BuyerLogonId,
		n.SellerId,
		n.SellerEmail,
		n.TradeStatus,
		n.TotalAmount,
		n.ReceiptAmount,
		n.InvoiceAmount,
		n.BuyerPayAmount,
		n.PointAmount,
		n.RefundFee,
		n.GmtCreate,
		n.GmtPayment,
		n.GmtRefund,
		n.GmtClose,
		n.FundBillList,
		n.PassbackParams,
		n.VoucherDetailList,
	)

	if err != nil {
		logger.WithField("trace", "SaveWxNotification").Error(err)
		return err
	}

	return nil
}
