package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/smartwalle/alipay"
)

// SaveAliNotification logs everything Alipay sends.
// Deprecated. Use SaveAliWebhookPayload
func (env Env) SaveAliNotification(n alipay.TradeNotification) error {

	_, err := env.dbs.Write.Exec(ali.StmtInsertAliPayLoad,
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

func (env Env) SaveAliWebhookPayload(p ali.WebhookPayload) error {
	_, err := env.dbs.Write.NamedExec(
		ali.StmtSavePayload,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) SaveWxPayload(schema wechat.PayloadSchema) error {
	_, err := env.dbs.Write.NamedExec(
		wechat.StmtSavePayload,
		schema)

	if err != nil {
		return err
	}

	return nil
}
