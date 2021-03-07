// +build !production

package invoice

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/cart"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/google/uuid"
	"github.com/guregu/null"
)

type MockInvoiceBuilder struct {
	id          string
	userID      string
	orderID     string
	price       price.FtcPrice
	orderKind   enum.OrderKind
	payMethod   enum.PayMethod
	addOnSource addon.Source
}

func NewMockInvoiceBuilder(userID string) MockInvoiceBuilder {
	if userID == "" {
		userID = uuid.New().String()
	}

	return MockInvoiceBuilder{
		id:          db.InvoiceID(),
		userID:      userID,
		orderID:     db.MustOrderID(),
		price:       faker.PriceStdYear,
		orderKind:   enum.OrderKindCreate,
		payMethod:   enum.PayMethodAli,
		addOnSource: "",
	}
}

func (b MockInvoiceBuilder) WithOrderID(id string) MockInvoiceBuilder {
	b.orderID = id
	return b
}

func (b MockInvoiceBuilder) WithPrice(p price.FtcPrice) MockInvoiceBuilder {
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

func (b MockInvoiceBuilder) WithAddOnSource(s addon.Source) MockInvoiceBuilder {
	b.addOnSource = s
	return b
}

func (b MockInvoiceBuilder) WithPayMethod(m enum.PayMethod) MockInvoiceBuilder {
	b.payMethod = m
	return b
}

func (b MockInvoiceBuilder) Build() Invoice {
	item := cart.NewFtcCart(b.price)

	return Invoice{
		ID:             b.id,
		CompoundID:     b.userID,
		Edition:        item.Price.Edition,
		YearMonthDay:   dt.NewYearMonthDay(item.Price.Cycle),
		AddOnSource:    b.addOnSource,
		OrderID:        null.StringFrom(b.orderID),
		OrderKind:      b.orderKind,
		PaidAmount:     item.Payable().Amount,
		PaymentMethod:  b.payMethod,
		PriceID:        null.StringFrom(item.Price.ID),
		CreatedUTC:     chrono.TimeNow(),
		ConsumedUTC:    chrono.Time{},
		DateTimePeriod: dt.DateTimePeriod{},
		CarriedOverUtc: chrono.Time{},
	}
}
