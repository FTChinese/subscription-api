//go:build !production

package ftcpay

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"github.com/smartwalle/alipay"
	"time"
)

type MockOrderBuilder struct {
	id        string
	ftcID     string
	unionID   string
	price     reader.PaywallPrice
	kind      enum.OrderKind
	payMethod enum.PayMethod
	wxAppId   null.String
	confirmed bool
	period    dt.SlotBuilder
	offerKind price.OfferKind
}

func NewMockOrderBuilder(ftcID string) MockOrderBuilder {

	return MockOrderBuilder{
		id:        ids.MustOrderID(),
		ftcID:     ftcID,
		unionID:   "",
		price:     reader.MockPwPriceStdYear,
		kind:      enum.OrderKindCreate,
		payMethod: enum.PayMethodAli,
		confirmed: false,
		offerKind: price.OfferKindNull,
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

func (b MockOrderBuilder) WithPrice(p reader.PaywallPrice) MockOrderBuilder {
	b.price = p
	return b
}

func (b MockOrderBuilder) WithStdYear() MockOrderBuilder {
	return b.WithPrice(reader.MockPwPriceStdYear)
}

func (b MockOrderBuilder) WithStdMonth() MockOrderBuilder {
	return b.WithPrice(reader.MockPwPriceStdMonth)
}

func (b MockOrderBuilder) WithPrm() MockOrderBuilder {
	return b.WithPrice(reader.MockPwPricePrm)
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

func (b MockOrderBuilder) WithOfferKinds(k price.OfferKind) MockOrderBuilder {
	b.offerKind = k
	return b
}

func (b MockOrderBuilder) WithPayMethod(m enum.PayMethod) MockOrderBuilder {
	b.payMethod = m
	if m == enum.PayMethodWx {
		b.wxAppId = null.StringFrom(faker.WxUnionID())
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
	b.period = dt.NewSlotBuilder(from).
		WithCycle(b.price.Cycle)

	return b
}

func (b MockOrderBuilder) findDiscount() price.Discount {
	for _, v := range b.price.Offers {
		if v.Kind == b.offerKind {
			return v
		}
	}

	return price.Discount{}
}
func (b MockOrderBuilder) Build() Order {

	item := reader.CartItemFtc{
		Price: b.price.FtcPrice,
		Offer: b.findDiscount(),
	}

	ymd := item.PeriodCount()

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
		Tier:          b.price.Tier,
		Cycle:         b.price.Cycle,
		Kind:          b.kind,
		OriginalPrice: b.price.UnitAmount,
		PayableAmount: item.PayableAmount(),
		PaymentMethod: b.payMethod,
		YearsCount:    ymd.Years,
		MonthsCount:   ymd.Months,
		DaysCount:     ymd.Days,
		WxAppID:       b.wxAppId,
		ConfirmedAt:   chrono.TimeFrom(confirmed),
		CreatedAt:     chrono.TimeNow(),
		StartDate:     b.period.StartDate(),
		EndDate:       b.period.EndDate(),
	}
}

func MockNewPaymentResult(o Order) PaymentResult {
	switch o.PaymentMethod {
	case enum.PayMethodWx:
		result := PaymentResult{
			PaymentState:     "SUCCESS",
			PaymentStateDesc: "",
			Amount:           null.IntFrom(o.WxPayable()),
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
			Amount:           null.IntFrom(o.WxPayable()),
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

func MockAliNoti(order Order) *alipay.TradeNotification {
	return &alipay.TradeNotification{
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
		TotalAmount:       order.AliPayable(),
		ReceiptAmount:     order.AliPayable(),
		InvoiceAmount:     order.AliPayable(),
		BuyerPayAmount:    order.AliPayable(),
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
