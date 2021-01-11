// +build !production

package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"

	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/smartwalle/alipay"
	"time"
)

func MockOrder(plan product.ExpandedPlan, kind enum.OrderKind) Order {
	id := uuid.New().String()
	return Order{
		ID: db.MustOrderID(),
		MemberID: reader.MemberID{
			CompoundID: id,
			FtcID:      null.StringFrom(id),
			UnionID:    null.String{},
		},
		PlanID:     plan.ID,
		DiscountID: plan.Discount.DiscID,
		Price:      plan.Price,
		Edition:    plan.Edition,
		Charge: product.Charge{
			Amount:   plan.Price,
			Currency: "cny",
		},
		Kind:          kind,
		PaymentMethod: enum.PayMethodAli,
		WxAppID:       null.String{},
		CreatedAt:     chrono.TimeNow(),
		ConfirmedAt:   chrono.Time{},
		DateRange:     dt.DateRange{},
		LiveMode:      true,
	}
}

func MockAliNoti(order Order) alipay.TradeNotification {
	return alipay.TradeNotification{
		AuthAppId:         "",
		NotifyTime:        time.Now().In(time.UTC).Format(chrono.SQLDateTime),
		NotifyType:        "trade_status_sync",
		NotifyId:          rand.String(36),
		AppId:             ali.MustInitApp().ID,
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
		ReceiptAmount:     order.AliPrice(),
		InvoiceAmount:     order.AliPrice(),
		BuyerPayAmount:    order.AliPrice(),
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
