// +build !production

package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"

	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/smartwalle/alipay"
	"time"
)

func MockBalanceSource() BalanceSource {
	return BalanceSource{
		OrderID:   MustGenerateOrderID(),
		Amount:    258.00,
		StartDate: chrono.DateFrom(time.Now()),
		EndDate:   chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
	}
}

func MockBalanceSourceN(n int) []BalanceSource {
	bs := make([]BalanceSource, 0)
	for i := 0; i < n; i++ {
		bs = append(bs, MockBalanceSource())
	}

	return bs
}

func MockProratedOrder() ProratedOrder {
	return ProratedOrder{
		OrderID:        MustGenerateOrderID(),
		Balance:        float64(rand.IntRange(10, 259)),
		CreatedUTC:     chrono.TimeNow(),
		ConsumedUTC:    chrono.Time{},
		UpgradeOrderID: "",
	}
}

func MockProratedOrderN(n int) []ProratedOrder {
	upID := MustGenerateOrderID()

	pos := make([]ProratedOrder, 0)

	for i := 0; i < n; i++ {
		o := MockProratedOrder()
		o.ConsumedUTC = chrono.TimeNow()
		o.UpgradeOrderID = upID

		pos = append(pos, o)
	}

	return pos
}

func MockOrder() Order {
	id := uuid.New().String()
	return Order{
		ID: MustGenerateOrderID(),
		MemberID: reader.MemberID{
			CompoundID: id,
			FtcID:      null.StringFrom(id),
			UnionID:    null.String{},
		},
		PlanID:     "",
		DiscountID: null.String{},
		Price:      faker.PlanStdYear.Price,
		Edition:    faker.PlanStdYear.Edition,
		Charge: product.Charge{
			Amount:   faker.PlanStdYear.Price,
			Currency: "cny",
		},
		Duration: product.Duration{
			CycleCount: 1,
			ExtraDays:  1,
		},
		Kind:            enum.OrderKindCreate,
		PaymentMethod:   enum.PayMethodAli,
		TotalBalance:    null.Float{},
		WxAppID:         null.String{},
		CreatedAt:       chrono.TimeNow(),
		ConfirmedAt:     chrono.Time{},
		PurchasedPeriod: PurchasedPeriod{},
		LiveMode:        true,
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
