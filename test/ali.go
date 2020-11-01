// +build !production

package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/smartwalle/alipay"
	"time"
)

func AliNoti(order subs.Order) alipay.TradeNotification {
	return alipay.TradeNotification{
		AuthAppId:         "",
		NotifyTime:        time.Now().In(time.UTC).Format(chrono.SQLDateTime),
		NotifyType:        "trade_status_sync",
		NotifyId:          rand.String(36),
		AppId:             AliApp.ID,
		Charset:           "utf-8",
		Version:           "1.0",
		SignType:          "RSA2",
		Sign:              rand.String(256),
		TradeNo:           rand.String(64),
		OutTradeNo:        order.ID,
		OutBizNo:          "",
		BuyerId:           "",
		BuyerLogonId:      "",
		SellerId:          "",
		SellerEmail:       "",
		TradeStatus:       "TRADE_SUCCESS",
		TotalAmount:       order.AliPrice(),
		ReceiptAmount:     "",
		InvoiceAmount:     "",
		BuyerPayAmount:    "",
		PointAmount:       "",
		RefundFee:         "",
		Subject:           "",
		Body:              "",
		GmtCreate:         time.Now().In(time.UTC).Format(chrono.SQLDateTime),
		GmtPayment:        time.Now().In(time.UTC).Format(chrono.SQLDateTime),
		GmtRefund:         "",
		GmtClose:          "",
		FundBillList:      "",
		PassbackParams:    "",
		VoucherDetailList: "",
	}
}
