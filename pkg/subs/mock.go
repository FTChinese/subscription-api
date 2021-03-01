// +build !production

package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/cart"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"

	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/smartwalle/alipay"
	"time"
)

type MockOrderBuilder struct {
	id        string
	userIDs   reader.MemberID
	price     price.FtcPrice
	kind      enum.OrderKind
	payMethod enum.PayMethod
	wxAppId   null.String
	confirmed bool
	period    dt.DatePeriod
}

func NewMockOrderBuilder(id string) MockOrderBuilder {
	if id == "" {
		id = db.MustOrderID()
	}

	return MockOrderBuilder{
		id: id,
		userIDs: reader.MemberID{
			CompoundID: "",
			FtcID:      null.StringFrom(uuid.New().String()),
			UnionID:    null.String{},
		}.MustNormalize(),
		price:     faker.PriceStdYear,
		kind:      enum.OrderKindCreate,
		payMethod: enum.PayMethodAli,
		confirmed: false,
	}
}

func (b MockOrderBuilder) WithUserIDs(ids reader.MemberID) MockOrderBuilder {
	b.userIDs = ids
	return b
}

func (b MockOrderBuilder) WithPrice(p price.FtcPrice) MockOrderBuilder {
	b.price = p
	return b
}

func (b MockOrderBuilder) WithKind(k enum.OrderKind) MockOrderBuilder {
	b.kind = k
	return b
}

func (b MockOrderBuilder) WithPayMethod(m enum.PayMethod) MockOrderBuilder {
	b.payMethod = m
	if m == enum.PayMethodWx {
		b.wxAppId = null.StringFrom(faker.GenWxID())
	}
	return b
}

func (b MockOrderBuilder) WithConfirmed() MockOrderBuilder {
	b.confirmed = true
	return b
}

func (b MockOrderBuilder) WithStartTime(from time.Time) MockOrderBuilder {
	if !b.confirmed {
		b.confirmed = true
	}
	b.period = dt.NewTimeRange(from).
		WithCycle(b.price.Cycle).
		ToDatePeriod()
	return b
}

func (b MockOrderBuilder) Build() Order {
	item := cart.NewFtcCart(b.price)

	payable := item.Payable()

	var confirmed time.Time
	if b.confirmed {
		confirmed = time.Now()
	}

	return Order{
		ID:            b.id,
		MemberID:      b.userIDs,
		PlanID:        item.Price.ID,
		DiscountID:    item.Discount.DiscID,
		Price:         item.Price.UnitAmount,
		Edition:       item.Price.Edition,
		Charge:        payable,
		Kind:          b.kind,
		PaymentMethod: b.payMethod,
		WxAppID:       b.wxAppId,
		CreatedAt:     chrono.TimeNow(),
		ConfirmedAt:   chrono.TimeFrom(confirmed),
		DatePeriod:    b.period,
		LiveMode:      true,
	}
}

func MockNewPaymentResult(o Order) PaymentResult {
	switch o.PaymentMethod {
	case enum.PayMethodWx:
		result := PaymentResult{
			PaymentState:     "SUCCESS",
			PaymentStateDesc: "",
			Amount:           null.IntFrom(o.AmountInCent()),
			TransactionID:    rand.String(28),
			OrderID:          o.ID,
			PaidAt:           chrono.TimeNow(),
			ConfirmedUTC:     chrono.TimeNow(),
			PayMethod:        enum.PayMethodWx,
		}
		return result

	case enum.PayMethodAli:
		return PaymentResult{
			PaymentState:     "TRADE_SUCCESS",
			PaymentStateDesc: "",
			Amount:           null.IntFrom(o.AmountInCent()),
			TransactionID:    rand.String(28),
			OrderID:          o.ID,
			PaidAt:           chrono.TimeNow(),
			ConfirmedUTC:     chrono.TimeNow(),
			PayMethod:        enum.PayMethodAli,
		}

	default:
		panic("Not ali or wx pay")
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
