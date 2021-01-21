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
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"

	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/smartwalle/alipay"
	"time"
)

func MockOrder(plan product.ExpandedPlan, kind enum.OrderKind) Order {
	return NewMockOrderBuilder("").
		WithPlan(plan).
		WithKind(kind).
		Build()
}

type MockOrderBuilder struct {
	id        string
	userIDs   reader.MemberID
	plan      product.ExpandedPlan
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
		plan:      faker.PlanStdYear,
		kind:      enum.OrderKindCreate,
		payMethod: enum.PayMethodAli,
		confirmed: false,
	}
}

func (b MockOrderBuilder) WithUserIDs(ids reader.MemberID) MockOrderBuilder {
	b.userIDs = ids
	return b
}

func (b MockOrderBuilder) WithPlan(p product.ExpandedPlan) MockOrderBuilder {
	b.plan = p
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
	b.period = dt.NewDateRange(from).WithCycle(b.plan.Cycle)
	return b
}

func (b MockOrderBuilder) Build() Order {
	item := NewCheckedItem(b.plan)

	payable := item.Payable()

	var confirmed time.Time
	if b.confirmed {
		confirmed = time.Now()
	}

	return Order{
		ID:         b.id,
		MemberID:   b.userIDs,
		PlanID:     item.Plan.ID,
		DiscountID: item.Discount.DiscID,
		Price:      item.Plan.Price,
		Edition:    item.Plan.Edition,
		Charge: product.Charge{
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
	plan      product.ExpandedPlan
	payMethod enum.PayMethod
}

func NewMockAddOnBuilder() MockAddOnBuilder {
	return MockAddOnBuilder{
		userIDs: reader.MemberID{
			FtcID: null.StringFrom(uuid.New().String()),
		}.MustNormalize(),
		plan:      faker.PlanStdYear,
		payMethod: enum.PayMethodAli,
	}
}

func (b MockAddOnBuilder) WithUserIDs(ids reader.MemberID) MockAddOnBuilder {
	b.userIDs = ids
	return b
}

func (b MockAddOnBuilder) WithPlan(p product.ExpandedPlan) MockAddOnBuilder {
	b.plan = p
	return b
}

func (b MockAddOnBuilder) BuildNew() AddOn {
	return AddOn{
		ID:                 db.AddOnID(),
		Edition:            b.plan.Edition,
		CycleCount:         1,
		DaysRemained:       1,
		IsUpgradeCarryOver: false,
		PaymentMethod:      b.payMethod,
		CompoundID:         b.userIDs.CompoundID,
		OrderID:            null.StringFrom(db.MustOrderID()),
		PlanID:             null.StringFrom(b.plan.ID),
		CreatedUTC:         chrono.TimeNow(),
		ConsumedUTC:        chrono.Time{},
	}
}

func (b MockAddOnBuilder) BuildUpgrade() AddOn {
	return AddOn{
		ID:                 db.AddOnID(),
		Edition:            faker.PlanStdYear.Edition,
		CycleCount:         0,
		DaysRemained:       int64(rand.IntRange(1, 367)),
		IsUpgradeCarryOver: false,
		PaymentMethod:      b.payMethod,
		CompoundID:         b.userIDs.CompoundID,
		OrderID:            null.StringFrom(db.MustOrderID()),
		PlanID:             null.StringFrom(faker.PlanStdYear.ID),
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
