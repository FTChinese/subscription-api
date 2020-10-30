package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/smartwalle/alipay"
)

// SaveAliNotification logs everything Alipay sends.
func (env Env) SaveAliNotification(n alipay.TradeNotification) error {

	_, err := env.db.Exec(InsertAliPayLoad,
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
		return err
	}

	return nil
}

// SavePrepayResp saves Wechat prepay response for future analysis.
func (env Env) SavePrepayResp(resp wechat.OrderResp) error {

	_, err := env.db.NamedExec(
		InsertWxPrepay,
		resp,
	)

	if err != nil {
		return err
	}

	return nil
}

// SaveWxNotification saves a wechat notification for logging purpose.
func (env Env) SaveWxNotification(n wechat.Notification) error {

	_, err := env.db.NamedExec(
		InsertWxPayLoad,
		n,
	)

	if err != nil {
		return err
	}

	return nil
}

// SaveWxQueryResp stores wechat pay query result to DB.
func (env Env) SaveWxQueryResp(resp wechat.OrderQueryResp) error {

	_, err := env.db.NamedExec(
		InsertWxQueryPayLoad,
		resp,
	)

	if err != nil {
		return err
	}

	return nil
}
