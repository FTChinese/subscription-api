package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
)

// PaymentConfig collects parameters to build an order.
// These are experimental refactoring.
type PaymentConfig struct {
	Account reader.FtcAccount // Required. Who is paying.
	Plan    price.Price       // Deprecated
	Price   pw.ProductPrice   // Required. What is purchased.
	Method  enum.PayMethod    // Optional if no payment is actually involved.
	WxAppID null.String
}

// NewPayment initializes a new payment session.
// Who and what to purchase are the minimal data required to start payment.
func NewPayment(account reader.FtcAccount, price pw.ProductPrice) PaymentConfig {
	return PaymentConfig{
		Account: account,
		Price:   price,
	}
}

func (c PaymentConfig) WithAlipay() PaymentConfig {
	c.Method = enum.PayMethodAli
	return c
}

func (c PaymentConfig) WithWxpay(app wechat.PayApp) PaymentConfig {
	c.Method = enum.PayMethodWx
	c.WxAppID = null.StringFrom(app.AppID)
	return c
}

// Checkout determines how a user should check out. This version
// allows all user to pay via alipay or wxpay, even if current membership is a valid stripe or iap.
// If Kind == OrderKindAddOn,
func (c PaymentConfig) checkout(m reader.Membership) (Checkout, error) {

	intents, err := NewCheckoutIntents(m, c.Price.Original.Edition)
	if err != nil {
		return Checkout{}, err
	}

	intent, err := intents.FindIntent(c.Method)
	if err != nil {
		return Checkout{}, err
	}

	item := NewCheckoutItem(c.Price)
	return Checkout{
		Kind:     intent.OrderKind,
		Item:     item,
		Payable:  item.Payable(),
		LiveMode: true,
	}.WithTest(c.Account.IsTest()), nil
}

// BuildOrder creates an Order based on a checkout action.
func (c PaymentConfig) order(checkout Checkout) (Order, error) {

	orderID, err := db.OrderID()
	if err != nil {
		return Order{}, err
	}

	return Order{
		ID:         orderID,
		MemberID:   c.Account.MemberID(),
		PlanID:     checkout.Item.Price.ID,
		DiscountID: checkout.Item.Discount.DiscID,
		Price:      checkout.Item.Price.UnitAmount,
		Edition:    checkout.Item.Price.Edition,
		Charge: price.Charge{
			Amount:   checkout.Payable.Amount,
			Currency: checkout.Payable.Currency,
		},
		Kind:          checkout.Kind,
		PaymentMethod: c.Method,
		WxAppID:       c.WxAppID,
		DateRange:     dt.DateRange{},
		CreatedAt:     chrono.TimeNow(),
		ConfirmedAt:   chrono.Time{},
		LiveMode:      checkout.LiveMode,
	}, nil
}

// BuildIntent builds new a new payment intent based on
// current membership status.
func (c PaymentConfig) BuildIntent(m reader.Membership) (PaymentIntent, error) {
	checkout, err := c.checkout(m)
	if err != nil {
		return PaymentIntent{}, err
	}

	order, err := c.order(checkout)
	if err != nil {
		return PaymentIntent{}, err
	}

	return PaymentIntent{
		Checkout: checkout,
		Order:    order,
	}, nil
}
