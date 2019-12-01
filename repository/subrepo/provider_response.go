package subrepo

import (
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/models/query"
	"gitlab.com/ftchinese/subscription-api/models/wechat"
)

// SaveAliNotification logs everything Alipay sends.
func (env SubEnv) SaveAliNotification(n alipay.TradeNotification) error {

	_, err := env.db.Exec(query.InsertAliPayLoad,
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
		logger.WithField("trace", "SubEnv.SaveAliNotification").Error(err)
		return err
	}

	return nil
}

// SavePrepayResp saves Wechat prepay response for future analysis.
func (env SubEnv) SavePrepayResp(resp wechat.UnifiedOrderResp) error {

	_, err := env.db.NamedExec(
		query.InsertWxPrepay,
		resp,
	)

	if err != nil {
		return err
	}

	return nil
}

// SaveWxNotification saves a wechat notification for logging purpose.
func (env SubEnv) SaveWxNotification(n wechat.Notification) error {

	_, err := env.db.NamedExec(
		query.InsertWxPayLoad,
		n,
	)

	if err != nil {
		return err
	}

	return nil
}

// SaveWxQueryResp stores wechat pay query result to DB.
func (env SubEnv) SaveWxQueryResp(resp wechat.OrderQueryResp) error {

	_, err := env.db.NamedExec(
		query.InsertWxQueryPayLoad,
		resp,
	)

	if err != nil {
		return err
	}

	return nil
}
