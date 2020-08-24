package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"github.com/sirupsen/logrus"
	"github.com/smartwalle/alipay"
)

// OrderBuilder is used to builder an order based
// on user's current membership, chosen plan, and current
// wallet balance.
//
// For upgrading, some extra steps needs to be performed:
// * Build upgrading balance and where those balance comes from
//
// For free upgrade, some extra steps needs to be performed:
// * the created order is also confirmed;
// * old membership updated;
type OrderBuilder struct {
	// Required fields.
	memberID reader.MemberID
	plan     product.ExpandedPlan
	live     bool // Use live webhook or sandbox.
	env      config.BuildConfig

	wallet Wallet // Only required if kind == SubsKindUpgrade.

	// Optional fields.
	ip              string              // For wxpay only
	method          enum.PayMethod      // Should be enum.PayMethodNull for free upgrade.
	wxAppID         null.String         // Only required for wechat pay.
	wxUnifiedParams wechat.UnifiedOrder // Only for wechat pay.

	// Calculated from previous fields.
	kind  enum.OrderKind
	order Order

	isBuilt bool // A flag to ensure all fields are properly initialized.
}

func NewOrderBuilder(id reader.MemberID) *OrderBuilder {
	return &OrderBuilder{
		memberID: id,
		isBuilt:  false,
	}
}

func (b *OrderBuilder) SetWxAppID(appID string) *OrderBuilder {
	b.wxAppID = null.StringFrom(appID)
	return b
}

func (b *OrderBuilder) GetReaderID() reader.MemberID {
	return b.memberID
}

func (b *OrderBuilder) SetEnvConfig(c config.BuildConfig) *OrderBuilder {
	b.env = c
	return b
}

func (b *OrderBuilder) SetPlan(p product.ExpandedPlan) *OrderBuilder {
	b.plan = p
	return b
}

func (b *OrderBuilder) SetPayMethod(m enum.PayMethod) *OrderBuilder {
	b.method = m
	return b
}

// SetUserIP that will be used to build wxpay params.
func (b *OrderBuilder) SetUserIP(ip string) *OrderBuilder {
	b.ip = ip
	return b
}

// SetWxParams sets the unified order wechat created
// so that we could build proper response to wechat pay.
func (b *OrderBuilder) SetWxParams(p wechat.UnifiedOrder) *OrderBuilder {
	b.wxUnifiedParams = p
	return b
}

// getWebHookURL determines which url to use.
// For local development, we need a weird combination:
// use production db layout while using sandbox url.
func (b *OrderBuilder) getWebHookURL() string {
	baseURL := "http://www.ftacademy.cn/api"
	if b.live {
		baseURL = baseURL + "/v1"
	} else {
		baseURL = baseURL + "/sandbox"
	}

	switch b.method {
	case enum.PayMethodAli:
		return baseURL + "/webhook/alipay"

	case enum.PayMethodWx:
		return baseURL + "/webhook/wxpay"

	default:
		return ""
	}
}

// ----------
// The following parameters need to query db.

// DeduceSubsKind determines the subscription's usage based on existing membership target plan.
// The plan field must be set before calling this method.
func (b *OrderBuilder) DeduceSubsKind(m reader.Membership) error {
	if b.plan.ID == "" {
		return ErrInvalidPlan
	}

	kind, err := m.AliWxSubsKind(b.plan.Edition)
	if err != nil {
		return err
	}

	b.kind = kind
	return nil
}

// GetSubsKind returns the AliWxSubsKind, or error if it is
// cannot be deduced.
// See errors returned from buildSubsKind.
func (b *OrderBuilder) GetSubsKind() enum.OrderKind {
	return b.kind
}

// SetWallet if this is an upgrade order.
func (b *OrderBuilder) SetWallet(w Wallet) *OrderBuilder {
	b.wallet = w
	return b
}

func (b *OrderBuilder) GetWallet() Wallet {
	return b.wallet
}

// Build calculates subscription kind, the amount to pay,
// billing cycles user purchased.
// See buildSubsKind form returned errors.
func (b *OrderBuilder) Build() error {

	if b.isBuilt {
		return nil
	}

	// Wallet should exist only for upgrading order.
	if b.kind != enum.OrderKindUpgrade {
		b.wallet = Wallet{}
	}

	orderID, err := GenerateOrderID()
	if err != nil {
		return err
	}
	// A zero wallet still produces valid Duration value,
	// which default to 1 cycle plus 1 day.
	duration := b.wallet.ConvertBalance(b.plan)
	charge := product.Charge{
		Amount: b.plan.Amount() - b.wallet.Balance,
		//Currency: b.plan.Currency,
	}
	// If balance is larger than a plan's price,
	// charged amount should be zero.
	if charge.Amount < 0 {
		charge.Amount = 0
	}

	// For sandbox change to a fixed amount
	if b.env.Sandbox() {
		charge.Amount = 0.01
	}

	b.order = Order{
		ID:            orderID,
		Price:         b.plan.Price,
		Charge:        charge,
		MemberID:      b.memberID,
		PlanID:        b.plan.ID,
		DiscountID:    b.plan.Discount.DiscID,
		Edition:       b.plan.Edition,
		Duration:      duration,
		Kind:          b.kind,
		PaymentMethod: b.method,
		TotalBalance:  null.NewFloat(b.wallet.Balance, b.wallet.Balance != 0),
		WxAppID:       b.wxAppID,
		StartDate:     chrono.Date{},
		EndDate:       chrono.Date{},
		CreatedAt:     chrono.TimeNow(),
		ConfirmedAt:   chrono.Time{},
		LiveMode:      !b.env.Sandbox(),
	}

	// After order is generated, we can now update wallet's
	// prorated orders, if this order is used for upgrade.
	if b.kind == enum.OrderKindUpgrade {
		b.wallet = b.wallet.WithUpgradeOrder(b.order)
	}

	b.isBuilt = true
	return nil
}

func (b *OrderBuilder) GetOrder() (Order, error) {
	if !b.isBuilt {
		err := b.Build()
		if err != nil {
			return Order{}, err
		}
	}

	return b.order, nil
}

// PaymentIntent tells client how to guide user to pay when user queries upgrade balance,
// or a  free upgrade attempt is not satisfied and
// payment is required.
func (b *OrderBuilder) PaymentIntent() (PaymentIntent, error) {
	if !b.isBuilt {
		err := b.Build()
		if err != nil {
			return PaymentIntent{}, err
		}
	}

	return PaymentIntent{
		Charge:   b.order.Charge,
		Duration: b.order.Duration,
		SubsKind: b.order.Kind,
		Wallet:   b.wallet,
		Plan: product.IntentPlan{
			Plan:   b.plan.Plan,
			Charge: b.order.Charge, // why it exists?
		},
	}, nil
}

// The following methods builds various query parameters
// send to payment provider.

func (b *OrderBuilder) AliAppPayParams() alipay.AliPayTradeAppPay {
	webHook := b.getWebHookURL()
	logrus.WithField("trace", "OrderBuilder.AliAppPayParams").Infof("Using web hook url: %s", webHook)

	return alipay.AliPayTradeAppPay{
		TradePay: alipay.TradePay{
			NotifyURL:   webHook,
			Subject:     b.plan.PaymentTitle(b.kind),
			OutTradeNo:  b.order.ID,
			TotalAmount: b.order.AliPrice(),
			ProductCode: ali.ProductCodeApp.String(),
			GoodsType:   "0",
		},
	}
}

func (b *OrderBuilder) AliDesktopPayParams(retURL string) alipay.AliPayTradePagePay {
	return alipay.AliPayTradePagePay{
		TradePay: alipay.TradePay{
			NotifyURL:   b.getWebHookURL(),
			ReturnURL:   retURL,
			Subject:     b.plan.PaymentTitle(b.kind),
			OutTradeNo:  b.order.ID,
			TotalAmount: b.order.AliPrice(),
			ProductCode: ali.ProductCodeWeb.String(),
			GoodsType:   "0",
		},
	}
}

func (b *OrderBuilder) AliWapPayParams(retURL string) alipay.AliPayTradeWapPay {
	return alipay.AliPayTradeWapPay{
		TradePay: alipay.TradePay{
			NotifyURL:   b.getWebHookURL(),
			ReturnURL:   retURL,
			Subject:     b.plan.PaymentTitle(b.kind),
			OutTradeNo:  b.order.ID,
			TotalAmount: b.order.AliPrice(),
			ProductCode: ali.ProductCodeWeb.String(),
			GoodsType:   "0",
		},
	}
}

func (b *OrderBuilder) WxpayParams() wxpay.Params {
	webHook := b.getWebHookURL()
	logrus.WithField("trace", "OrderBuilder.AliAppPayParams").Infof("Using web hook url: %s", webHook)

	p := make(wxpay.Params)
	p.SetString("body", b.plan.PaymentTitle(b.kind))
	p.SetString("out_trade_no", b.order.ID)
	p.SetInt64("total_fee", b.order.AmountInCent())
	p.SetString("spbill_create_ip", b.ip)
	p.SetString("notify_url", webHook)
	// APP for native app
	// NATIVE for web site
	// JSAPI for web page opend inside wechat browser
	p.SetString("trade_type", b.wxUnifiedParams.TradeType.String())

	switch b.wxUnifiedParams.TradeType {
	case wechat.TradeTypeDesktop:
		p.SetString("product_id", b.plan.NamedKey())

	case wechat.TradeTypeJSAPI:
		p.SetString("openid", b.wxUnifiedParams.OpenID)
	}

	return p
}
