package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"time"
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

// Checkout is intermediate bridge between payment request and the final result.
type Checkout struct {
	Kind     enum.OrderKind   `json:"kind"`
	Item     CheckedItem      `json:"item"`
	Wallet   Wallet           `json:"wallet"`
	Duration product.Duration `json:"duration"`
	Payable  product.Charge   `json:"payable"`
	IsFree   bool             `json:"isFree"`
	LiveMode bool             `json:"live"`
}

func (c Checkout) WithTest(t bool) Checkout {
	c.LiveMode = !t

	if c.IsFree {
		return c
	}

	if t {
		c.Payable.Amount = 0.01
	}

	return c
}

// PaymentConfig collects parameters to build an order.
// These are experimental refactoring.
type PaymentConfig struct {
	dryRun  bool                 // Only for upgrade preview.
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

func (c PaymentConfig) WithUpgrade(preview bool) PaymentConfig {
	c.dryRun = preview
	return c
}

func (c PaymentConfig) Checkout(bs []BalanceSource, kind enum.OrderKind) Checkout {

	item := NewCheckedItem(c.Plan)

	w := Wallet{}
	if kind == enum.OrderKindUpgrade {
		w = NewWallet(bs, time.Now())
	}

	duration := w.ConvertLength(item.Plan)

	amount := item.Plan.Price - item.Discount.PriceOff.Float64 - w.Balance

	// Balance is large enough to cover price.
	if amount < 0 {
		return Checkout{
			Kind:     kind,
			Item:     item,
			Wallet:   w,
			Duration: duration,
			Payable: product.Charge{
				Amount:   0,
				Currency: "cny",
			},
			IsFree:   true,
			LiveMode: true,
		}.WithTest(c.Account.IsTest())
	}

	return Checkout{
		Kind:   kind,
		Item:   item,
		Wallet: w,
		Payable: product.Charge{
			Amount:   amount,
			Currency: "cny",
		},
		Duration: duration,
		IsFree:   false,
		LiveMode: true,
	}.WithTest(c.Account.IsTest())
}

func (c PaymentConfig) BuildOrder(checkout Checkout) (Order, error) {

	orderID, err := GenerateOrderID()
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
		Duration:        checkout.Duration,
		Kind:            checkout.Kind,
		PaymentMethod:   c.Method,
		TotalBalance:    null.NewFloat(checkout.Wallet.Balance, checkout.Wallet.Balance != 0),
		WxAppID:         c.WxAppID,
		PurchasedPeriod: PurchasedPeriod{},
		CreatedAt:       chrono.TimeNow(),
		ConfirmedAt:     chrono.Time{},
		LiveMode:        checkout.LiveMode,
	}, nil
}

func (c PaymentConfig) BuildIntent(bs []BalanceSource, kind enum.OrderKind) (PaymentIntent, error) {
	checkout := c.Checkout(bs, kind)
	order, err := c.BuildOrder(checkout)
	if err != nil {
		return PaymentIntent{}, err
	}

	return PaymentIntent{
		Checkout: checkout,
		Order:    order,
	}, nil
}

func (c PaymentConfig) UpgradeIntent(checkout Checkout, m reader.Membership) (UpgradeIntent, error) {

	intent := UpgradeIntent{
		Charge:         checkout.Payable,
		Duration:       checkout.Duration,
		LegacySubsKind: enum.OrderKindUpgrade,
		SubsKind:       enum.OrderKindUpgrade,
		Plan:           c.Plan,
		Discount:       checkout.Item.Discount,
		Wallet:         checkout.Wallet,
		Payable:        checkout.Payable,
		Length:         checkout.Duration,
		IsFree:         checkout.IsFree,
		Result:         ConfirmationResult{},
	}

	// This order is used for upgrade but the balance is not enough to cover the cost.
	if c.dryRun || !checkout.IsFree {
		return intent, nil
	}

	order, err := c.BuildOrder(checkout)
	if err != nil {
		return UpgradeIntent{}, err
	}

	confirmedAt := chrono.TimeNow()
	period, err := NewPeriodBuilder(
		c.Plan.Edition,
		checkout.Duration,
	).Build(chrono.DateNow())
	if err != nil {
		return UpgradeIntent{}, err
	}

	order.ConfirmedAt = confirmedAt
	order.PurchasedPeriod = period

	newMember, err := order.Membership()
	if err != nil {
		return UpgradeIntent{}, err
	}

	intent.Result = ConfirmationResult{
		Order:      order,
		Membership: newMember,
		Snapshot:   m.Snapshot(reader.FtcArchiver(order.Kind)),
	}

	return intent, nil
}
