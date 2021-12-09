//go:build !production
// +build !production

package invoice

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"time"
)

type MockInvoiceBuilder struct {
	id          string
	userID      string
	orderID     string
	price       pw.PaywallPrice
	orderKind   enum.OrderKind
	payMethod   enum.PayMethod
	addOnSource addon.Source
	startTime   time.Time
	offerKinds  []price.OfferKind
}

func NewMockInvoiceBuilder() MockInvoiceBuilder {

	return MockInvoiceBuilder{
		id:          ids.InvoiceID(),
		userID:      uuid.New().String(),
		orderID:     ids.MustOrderID(),
		price:       pw.MockPwPriceStdYear,
		orderKind:   enum.OrderKindCreate,
		payMethod:   enum.PayMethodAli,
		addOnSource: "",
		offerKinds: []price.OfferKind{
			price.OfferKindPromotion,
		},
	}
}

func (b MockInvoiceBuilder) WithFtcID(ftcID string) MockInvoiceBuilder {
	b.userID = ftcID
	return b
}

func (b MockInvoiceBuilder) WithUnionID(unionID string) MockInvoiceBuilder {
	b.userID = unionID
	return b
}

func (b MockInvoiceBuilder) WithOrderID(id string) MockInvoiceBuilder {
	b.orderID = id
	return b
}

func (b MockInvoiceBuilder) WithPrice(p pw.PaywallPrice) MockInvoiceBuilder {
	b.price = p
	return b
}

func (b MockInvoiceBuilder) WithOrderKind(k enum.OrderKind) MockInvoiceBuilder {
	b.orderKind = k
	if k == enum.OrderKindAddOn {
		b.addOnSource = addon.SourceUserPurchase
	}
	return b
}

func (b MockInvoiceBuilder) WithOfferKinds(k []price.OfferKind) MockInvoiceBuilder {
	b.offerKinds = k
	return b
}

func (b MockInvoiceBuilder) WithAddOnSource(s addon.Source) MockInvoiceBuilder {
	b.addOnSource = s
	if s == addon.SourceCompensation {
		b.orderID = ""
	}
	return b
}

func (b MockInvoiceBuilder) WithPayMethod(m enum.PayMethod) MockInvoiceBuilder {
	b.payMethod = m
	return b
}

func (b MockInvoiceBuilder) SetPeriodStart(t time.Time) MockInvoiceBuilder {
	b.startTime = t
	return b
}

func (b MockInvoiceBuilder) Build() Invoice {
	charge := price.NewCharge(b.price.Price, b.price.Offers.FindApplicable(b.offerKinds))

	if b.addOnSource != "" {
		b.orderKind = enum.OrderKindAddOn
	}

	return Invoice{
		ID:             b.id,
		CompoundID:     b.userID,
		Edition:        b.price.Edition,
		YearMonthDay:   dt.NewYearMonthDay(b.price.Cycle),
		AddOnSource:    b.addOnSource,
		AppleTxID:      null.String{},
		OrderID:        null.NewString(b.orderID, b.orderID != ""),
		OrderKind:      b.orderKind,
		PaidAmount:     charge.Amount,
		PaymentMethod:  b.payMethod,
		PriceID:        null.StringFrom(b.price.ID),
		StripeSubsID:   null.String{},
		CreatedUTC:     chrono.TimeNow(),
		ConsumedUTC:    chrono.Time{},
		DateTimePeriod: dt.DateTimePeriod{},
		CarriedOverUtc: chrono.Time{},
	}.SetPeriod(b.startTime)
}
