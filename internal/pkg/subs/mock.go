//go:build !production
// +build !production

package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"github.com/smartwalle/alipay"
	"time"
)

type MockOrderBuilder struct {
	id         string
	ftcID      string
	unionID    string
	price      price.FtcPrice
	kind       enum.OrderKind
	payMethod  enum.PayMethod
	wxAppId    null.String
	confirmed  bool
	period     dt.DatePeriod
	offerKinds []price.OfferKind
}

func NewMockOrderBuilder(ftcID string) MockOrderBuilder {

	return MockOrderBuilder{
		id:        ids.MustOrderID(),
		ftcID:     ftcID,
		unionID:   "",
		price:     price.MockPriceStdYear,
		kind:      enum.OrderKindCreate,
		payMethod: enum.PayMethodAli,
		confirmed: false,
		offerKinds: []price.OfferKind{
			price.OfferKindPromotion,
		},
	}
}

func (b MockOrderBuilder) WithFtcID(id string) MockOrderBuilder {
	b.ftcID = id
	return b
}

func (b MockOrderBuilder) WithUnionID(id string) MockOrderBuilder {
	b.unionID = id
	return b
}

func (b MockOrderBuilder) WithPrice(p price.FtcPrice) MockOrderBuilder {
	b.price = p
	return b
}

func (b MockOrderBuilder) WithStdYear() MockOrderBuilder {
	return b.WithPrice(price.MockPriceStdYear)
}

func (b MockOrderBuilder) WithStdMonth() MockOrderBuilder {
	return b.WithPrice(price.MockPriceStdMonth)
}

func (b MockOrderBuilder) WithPrm() MockOrderBuilder {
	return b.WithPrice(price.MockPricePrm)
}

func (b MockOrderBuilder) WithKind(k enum.OrderKind) MockOrderBuilder {
	b.kind = k
	return b
}

func (b MockOrderBuilder) WithCreate() MockOrderBuilder {
	return b.WithKind(enum.OrderKindCreate)
}

func (b MockOrderBuilder) WithRenew() MockOrderBuilder {
	return b.WithKind(enum.OrderKindRenew)
}

func (b MockOrderBuilder) WithAddOn() MockOrderBuilder {
	return b.WithKind(enum.OrderKindAddOn)
}

func (b MockOrderBuilder) WithUpgrade() MockOrderBuilder {
	return b.WithKind(enum.OrderKindUpgrade)
}

func (b MockOrderBuilder) WithOfferKinds(k []price.OfferKind) MockOrderBuilder {
	b.offerKinds = k
	return b
}

func (b MockOrderBuilder) WithPayMethod(m enum.PayMethod) MockOrderBuilder {
	b.payMethod = m
	if m == enum.PayMethodWx {
		b.wxAppId = null.StringFrom(faker.GenWxID())
	}
	return b
}

func (b MockOrderBuilder) WithAlipay() MockOrderBuilder {
	return b.WithPayMethod(enum.PayMethodAli)
}

func (b MockOrderBuilder) WithWx() MockOrderBuilder {
	return b.WithPayMethod(enum.PayMethodWx)
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

	discount := b.price.Offers.FindApplicable(b.offerKinds)
	charge := price.NewCharge(b.price.Price, discount)

	var confirmed time.Time
	if b.confirmed {
		confirmed = time.Now()
	}

	return Order{
		ID: b.id,
		UserIDs: ids.UserIDs{
			CompoundID: "",
			FtcID:      null.StringFrom(b.ftcID),
			UnionID:    null.NewString(b.unionID, b.unionID != ""),
		}.MustNormalize(),
		PlanID:        b.price.ID,
		DiscountID:    null.NewString(discount.ID, discount.ID != ""),
		Price:         b.price.UnitAmount,
		Edition:       b.price.Edition,
		Charge:        charge,
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
