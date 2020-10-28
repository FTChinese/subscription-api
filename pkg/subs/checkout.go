package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"time"
)

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

func (c Checkout) ProratedOrders(upgradeTo Order) []ProratedOrder {
	if c.Kind != enum.OrderKindUpgrade {
		return nil
	}
	if c.Wallet.Sources == nil || len(c.Wallet.Sources) == 0 {
		return nil
	}

	now := chrono.TimeNow()

	for i, v := range c.Wallet.Sources {
		v.UpgradeOrderID = upgradeTo.ID
		if c.IsFree {
			v.ConsumedUTC = now
		}

		c.Wallet.Sources[i] = v
	}

	return c.Wallet.Sources
}

// PaymentConfig collects parameters to build an order.
// These are experimental refactoring.
type PaymentConfig struct {
	dryRun         bool // Only for upgrade preview.
	Account        reader.FtcAccount
	Plan           product.ExpandedPlan
	Method         enum.PayMethod
	WebhookBaseURL string

	WxAppID null.String
}

// NewUpgradeConfig is used to build an upgrade intent, or
// free upgrade. It does not involve any payment providers.
func NewUpgradeConfig(a reader.FtcAccount, plan product.ExpandedPlan) PaymentConfig {
	return PaymentConfig{
		Account: a,
		Plan:    plan,
	}
}

func (c PaymentConfig) WithPreview(p bool) PaymentConfig {
	c.dryRun = p
	return c
}

func (c PaymentConfig) WebhookURL() string {
	switch c.Method {
	case enum.PayMethodAli:
		return c.WebhookBaseURL + "/webhook/alipay"

	case enum.PayMethodWx:
		return c.WebhookBaseURL + "/webhook/wxpay"

	default:
		return ""
	}
}

func (c PaymentConfig) Checkout(bs []BalanceSource, kind enum.OrderKind) Checkout {

	item := NewCheckedItem(c.Plan)

	w := Wallet{}
	if kind == enum.OrderKindUpgrade {
		w = NewWallet(bs, time.Now())
	}

	period := w.ConvertLength(item.Plan)

	amount := item.Plan.Price - item.Discount.PriceOff.Float64 - w.Balance

	// Balance is large enough to cover price.
	if amount < 0 {
		return Checkout{
			Kind:   kind,
			Wallet: w,
			Payable: product.Charge{
				Amount:   0,
				Currency: "cny",
			},
			Duration: period,
			IsFree:   true,
		}.WithTest(c.Account.IsTest())
	}

	return Checkout{
		Kind:   kind,
		Wallet: w,
		Payable: product.Charge{
			Amount:   amount,
			Currency: "cny",
		},
		Duration: period,
		IsFree:   false,
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
		CheckedItem:     checkout.Item,
		WebhookURL:      c.WebhookURL(),
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
