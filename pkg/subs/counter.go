package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/cart"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
)

// Counter is where user present shopping cart and wallet to check out.
type Counter struct {
	Account reader.FtcAccount // Required. Who is paying.
	Price   price.FtcPrice    // Required. What is purchased.
	Method  enum.PayMethod    // Optional if no payment is actually involved.
	WxAppID null.String
}

// NewCounter initializes a new payment session.
// Who and what to purchase are the minimal data required to start payment.
func NewCounter(account reader.FtcAccount, price price.FtcPrice) Counter {
	return Counter{
		Account: account,
		Price:   price,
	}
}

func (c Counter) WithAlipay() Counter {
	c.Method = enum.PayMethodAli
	return c
}

func (c Counter) WithWxpay(app wechat.PayApp) Counter {
	c.Method = enum.PayMethodWx
	c.WxAppID = null.StringFrom(app.AppID)
	return c
}

// Checkout determines how a user should check out. This version
// allows all user to pay via alipay or wxpay, even if current membership is a valid stripe or iap.
func (c Counter) checkout(m reader.Membership) (Checkout, error) {

	intent, err := cart.NewCheckoutIntents(m, c.Price.Edition).
		Get(c.Method)
	if err != nil {
		return Checkout{}, err
	}

	ftcCart := cart.NewFtcCart(c.Price)
	return Checkout{
		Kind:     intent.OneTimeKind,
		Cart:     ftcCart,
		Payable:  ftcCart.Payable(),
		LiveMode: true,
	}.WithTest(c.Account.IsTest()), nil
}

// BuildOrder creates an Order based on a checkout action.
func (c Counter) BuildOrder(m reader.Membership) (Order, error) {

	checkout, err := c.checkout(m)
	if err != nil {
		return Order{}, err
	}

	orderID, err := db.OrderID()
	if err != nil {
		return Order{}, err
	}

	return Order{
		ID:            orderID,
		MemberID:      c.Account.MemberID(),
		PlanID:        checkout.Cart.Price.ID,
		DiscountID:    checkout.Cart.Discount.DiscID,
		Price:         checkout.Cart.Price.UnitAmount,
		Edition:       checkout.Cart.Price.Edition,
		Charge:        checkout.Payable,
		Kind:          checkout.Kind,
		PaymentMethod: c.Method,
		WxAppID:       c.WxAppID,
		DatePeriod:    dt.DatePeriod{},
		CreatedAt:     chrono.TimeNow(),
		ConfirmedAt:   chrono.Time{},
		LiveMode:      checkout.LiveMode,
	}, nil
}
