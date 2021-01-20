package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
)

func WebhookURL(sandbox bool, method enum.PayMethod) string {
	var baseURL string
	if sandbox {
		baseURL = "http://www.ftacademy.cn/api/sandbox"
	} else {
		baseURL = "http://www.ftacademy.cn/api/v1"
	}

	switch method {
	case enum.PayMethodWx:
		return baseURL + "/webhook/wxpay"
	case enum.PayMethodAli:
		return baseURL + "/webhook/alipay"
	}

	return ""
}

// CheckedItem contains an item user want to buy and all attributes attached to it like applicable discount, coupon, etc..
type CheckedItem struct {
	Plan     product.Plan     `json:"plan"`
	Discount product.Discount `json:"discount"`
}

func NewCheckedItem(ep product.ExpandedPlan) CheckedItem {
	if ep.Discount.IsValid() {
		return CheckedItem{
			Plan:     ep.Plan,
			Discount: ep.Discount,
		}
	}

	return CheckedItem{
		Plan:     ep.Plan,
		Discount: product.Discount{},
	}
}

// Amount calculates the actual amount user should pay for a plan,
// after taking into account applicable discount, coupon, limited time offer, etc..
func (i CheckedItem) Payable() product.Charge {
	return product.Charge{
		Amount:   i.Plan.Price - i.Discount.PriceOff.Float64,
		Currency: "cny",
	}
}

// Checkout is intermediate bridge between payment request and the final result.
type Checkout struct {
	Kind     enum.OrderKind `json:"kind"`
	Item     CheckedItem    `json:"item"`
	Payable  product.Charge `json:"payable"`
	LiveMode bool           `json:"live"`
}

func (c Checkout) WithTest(t bool) Checkout {
	c.LiveMode = !t

	if t {
		c.Payable.Amount = 0.01
	}

	return c
}

// PaymentConfig collects parameters to build an order.
// These are experimental refactoring.
type PaymentConfig struct {
	Account reader.FtcAccount    // Required. Who is paying.
	Plan    product.ExpandedPlan // Required. What is purchased.
	Method  enum.PayMethod       // Optional if no payment is actually involved.
	WxAppID null.String
}

// NewPayment initializes a new payment session.
// Who and what to purchase are the minimal data required to start payment.
func NewPayment(account reader.FtcAccount, plan product.ExpandedPlan) PaymentConfig {
	return PaymentConfig{
		Account: account,
		Plan:    plan,
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

	kind, err := m.AliWxSubsKind(c.Plan.Edition)
	if err != nil {
		return Checkout{}, err
	}

	item := NewCheckedItem(c.Plan)
	return Checkout{
		Kind:     kind,
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
		PlanID:     checkout.Item.Plan.ID,
		DiscountID: checkout.Item.Discount.DiscID,
		Price:      checkout.Item.Plan.Price,
		Edition:    checkout.Item.Plan.Edition,
		Charge: product.Charge{
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
