package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
)

// Counter is where user present shopping cart and wallet to check out.
type Counter struct {
	BaseAccount account.BaseAccount // Required. Who is paying
	FtcPrice    price.FtcPrice      // Required. What is purchased.
	Method      enum.PayMethod      // Optional if no payment is actually involved.
	WxAppID     null.String
}

// NewCounter initializes a new payment session.
// Who and what to purchase are the minimal data required to start payment.
func NewCounter(a account.BaseAccount, price price.FtcPrice) Counter {
	return Counter{
		BaseAccount: a,
		FtcPrice:    price,
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

func (c Counter) order(checkout Checkout) (Order, error) {
	orderID, err := pkg.OrderID()
	if err != nil {
		return Order{}, err
	}

	return Order{
		ID:            orderID,
		UserIDs:       c.BaseAccount.CompoundIDs(),
		PlanID:        checkout.Price.ID,
		DiscountID:    checkout.Offer.DiscID,
		Price:         checkout.Price.UnitAmount,
		Edition:       checkout.Price.Edition,
		Charge:        checkout.Payable(),
		Kind:          checkout.Kind,
		PaymentMethod: c.Method,
		WxAppID:       c.WxAppID,
		DatePeriod:    dt.DatePeriod{},
		CreatedAt:     chrono.TimeNow(),
		ConfirmedAt:   chrono.Time{},
		LiveMode:      checkout.LiveMode,
	}, nil
}

func (c Counter) PaymentIntent(m reader.Membership) (PaymentIntent, error) {
	checkout, err := NewCheckout(c.FtcPrice, m)
	if err != nil {
		return PaymentIntent{}, err
	}

	checkout = checkout.WithTest(c.BaseAccount.IsTest())

	order, err := c.order(checkout)

	return PaymentIntent{
		Pricing: checkout.Price,
		Offer:   checkout.Offer,
		Order:   order,
	}, nil
}
