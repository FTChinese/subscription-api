// +build !production

package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/pkg/reader"

	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/smartwalle/alipay"
	"time"
)

func MockOrder(price pw.ProductPrice, kind enum.OrderKind) Order {
	return NewMockOrderBuilder("").
		WithPrice(price).
		WithKind(kind).
		Build()
}

type MockOrderBuilder struct {
	id        string
	userIDs   reader.MemberID
	price     pw.ProductPrice
	kind      enum.OrderKind
	payMethod enum.PayMethod
	wxAppId   null.String
	confirmed bool
	period    dt.DateRange
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

func (b MockOrderBuilder) WithPrice(p pw.ProductPrice) MockOrderBuilder {
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

func (b MockOrderBuilder) WithPeriod(from time.Time) MockOrderBuilder {
	b.period = dt.NewDateRange(from).WithCycle(b.price.Original.Cycle)
	return b
}

func (b MockOrderBuilder) Build() Order {
	item := NewCheckedItem(b.price)

	payable := item.Payable()

	var confirmed time.Time
	if b.confirmed {
		confirmed = time.Now()
	}

	return Order{
		ID:         b.id,
		MemberID:   b.userIDs,
		PlanID:     item.Price.ID,
		DiscountID: item.Discount.DiscID,
		Price:      item.Price.UnitAmount,
		Edition:    item.Price.Edition,
		Charge: price.Charge{
			Amount:   payable.Amount,
			Currency: payable.Currency,
		},
		Kind:          b.kind,
		PaymentMethod: b.payMethod,
		WxAppID:       b.wxAppId,
		CreatedAt:     chrono.TimeNow(),
		ConfirmedAt:   chrono.TimeFrom(confirmed),
		DateRange:     dt.DateRange{},
		LiveMode:      true,
	}
}

type MockAddOnBuilder struct {
	userIDs   reader.MemberID
	price     pw.ProductPrice
	payMethod enum.PayMethod
}

func NewMockAddOnBuilder() MockAddOnBuilder {
	return MockAddOnBuilder{
		userIDs: reader.MemberID{
			FtcID: null.StringFrom(uuid.New().String()),
		}.MustNormalize(),
		price:     faker.PriceStdYear,
		payMethod: enum.PayMethodAli,
	}
}

func (b MockAddOnBuilder) WithUserIDs(ids reader.MemberID) MockAddOnBuilder {
	b.userIDs = ids
	return b
}

func (b MockAddOnBuilder) WithPlan(p pw.ProductPrice) MockAddOnBuilder {
	b.price = p
	return b
}

func (b MockAddOnBuilder) BuildNew() AddOn {
	return AddOn{
		ID:                 db.AddOnID(),
		Edition:            b.price.Original.Edition,
		CycleCount:         1,
		DaysRemained:       1,
		IsUpgradeCarryOver: false,
		PaymentMethod:      b.payMethod,
		CompoundID:         b.userIDs.CompoundID,
		OrderID:            null.StringFrom(db.MustOrderID()),
		PlanID:             null.StringFrom(b.price.Original.ID),
		CreatedUTC:         chrono.TimeNow(),
		ConsumedUTC:        chrono.Time{},
	}
}

func (b MockAddOnBuilder) BuildUpgrade() AddOn {
	return AddOn{
		ID:                 db.AddOnID(),
		Edition:            faker.PriceStdYear.Original.Edition,
		CycleCount:         0,
		DaysRemained:       int64(rand.IntRange(1, 367)),
		IsUpgradeCarryOver: false,
		PaymentMethod:      b.payMethod,
		CompoundID:         b.userIDs.CompoundID,
		OrderID:            null.StringFrom(db.MustOrderID()),
		PlanID:             null.StringFrom(faker.PriceStdYear.Original.ID),
		CreatedUTC:         chrono.TimeNow(),
		ConsumedUTC:        chrono.Time{},
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
