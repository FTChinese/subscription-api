package subscription

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/models/ali"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/models/wechat"
)

type OrderBuilder struct {
	memberID reader.MemberID
	method   enum.PayMethod
	wxAppID  string
	plan     plan.Plan
	client util.ClientApp
	wxUnifiedParams wechat.UnifiedOrder

	// Requires querying DB
	wallet     Wallet
	membership Membership

	// Calculated for previous fields set.
	kind            plan.SubsKind
	charge          Charge
	duration        plan.Duration
	orderID         string
	upgradeSchemaID string

	isBuilt bool // A flag to ensure all fields are property initialized.
	sandbox bool // Whether this is under sandbox.
}

func NewOrderBuilder(id reader.MemberID) *OrderBuilder {
	return &OrderBuilder{
		memberID: id,
		duration: plan.Duration{
			CycleCount: 1,
			ExtraDays:  1,
		},
		isBuilt: false,
	}
}

func (b *OrderBuilder) SetWxAppID(appID string) *OrderBuilder {
	b.wxAppID = appID
	return b
}

func (b *OrderBuilder) GetReaderID() reader.MemberID {
	return b.memberID
}

func (b *OrderBuilder) SetSandbox() *OrderBuilder {
	b.sandbox = true
	return b
}

func (b *OrderBuilder) SetPlan(p plan.Plan) *OrderBuilder {
	b.plan = p
	return b
}

func (b *OrderBuilder) SetPayMethod(m enum.PayMethod) *OrderBuilder {
	b.method = m
	return b
}

func (b *OrderBuilder) SetClient(c util.ClientApp) *OrderBuilder {
	b.client = c
	return b
}

func (b *OrderBuilder) SetWxParams(p wechat.UnifiedOrder) *OrderBuilder {
	b.wxUnifiedParams = p
	return b
}

func (b *OrderBuilder) getWebHookURL() string {
	baseURL := "http://www.ftacademy.cn/api"
	if b.sandbox {
		baseURL = baseURL + "/sandbox"
	} else {
		baseURL = baseURL + "/v1"
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

func (b *OrderBuilder) generateOrderID() (err error) {
	if b.orderID != "" {
		return
	}

	b.orderID, err = GenerateOrderID()
	return
}

// ----------
// The following parameters need to query db.

func (b *OrderBuilder) SetMembership(m Membership) *OrderBuilder {
	b.membership = m

	return b
}

func (b *OrderBuilder) buildSubsKind() (err error) {
	if b.kind != plan.SubsKindNull {
		return
	}

	b.kind, err = b.membership.SubsKind(b.plan)

	if err != nil {
		return
	}

	return
}

func (b *OrderBuilder) GetSubsKind() (plan.SubsKind, error) {
	if err := b.buildSubsKind(); err != nil {
		return plan.SubsKindNull, err
	}

	return b.kind, nil
}

// SetBalance if this is an upgrade order.
func (b *OrderBuilder) SetWallet(w Wallet) *OrderBuilder {
	b.wallet = w
	return b
}

func (b *OrderBuilder) Build() error {

	if b.isBuilt {
		return nil
	}

	// Deduce subscription kind if not performed yet.
	if err := b.buildSubsKind(); err != nil {
		return err
	}

	b.duration = b.wallet.ConvertBalance(b.plan)
	b.charge = Charge{
		Amount:   b.plan.Price - b.wallet.Balance,
		Currency: b.plan.Currency,
	}

	if b.kind == plan.SubsKindUpgrade && b.upgradeSchemaID == "" {
		b.upgradeSchemaID = GenerateUpgradeID()
	}

	b.isBuilt = true
	return nil
}

func (b *OrderBuilder) ensureBuilt() error {
	if b.isBuilt {
		return nil
	}

	if err := b.Build(); err != nil {
		return err
	}

	return nil
}

func (b *OrderBuilder) PaymentIntent() (PaymentIntent, error) {
	if err := b.ensureBuilt(); err != nil {
		return PaymentIntent{}, err
	}

	return PaymentIntent{
		Charge:   b.charge,
		Duration: b.duration,
		SubsKind: b.kind,
		Wallet:   b.wallet,
		Plan:     b.plan,
	}, nil
}

func (b *OrderBuilder) Order() (Order, error) {

	if err := b.ensureBuilt(); err != nil {
		return Order{}, err
	}

	if err := b.generateOrderID(); err != nil {
		return Order{}, err
	}

	return Order{
		ID:               b.orderID,
		MemberID:         b.memberID,
		Tier:             b.plan.Tier,
		Cycle:            b.plan.Cycle,
		Price:            b.plan.Price,
		Amount:           b.charge.Amount,
		Currency:         b.plan.Currency,
		CycleCount:       b.duration.CycleCount,
		ExtraDays:        b.duration.ExtraDays,
		Usage:            b.kind,
		PaymentMethod:    b.method,
		WxAppID:          null.NewString(b.wxAppID, b.wxAppID != ""),
		UpgradeSchemaID:  null.NewString(b.upgradeSchemaID, b.upgradeSchemaID != ""),
		StartDate:        chrono.Date{},
		EndDate:          chrono.Date{},
		CreatedAt:        chrono.TimeNow(),
		ConfirmedAt:      chrono.Time{},
		MemberSnapshotID: null.String{}, // Set when confirming order.
	}, nil
}

func (b *OrderBuilder) UpgradeSchema() (UpgradeSchema, error) {

	if err := b.ensureBuilt(); err != nil {
		return UpgradeSchema{}, err
	}

	if b.kind != plan.SubsKindUpgrade {
		return UpgradeSchema{}, errors.New("not an upgrade subscription")
	}

	return UpgradeSchema{
		ID:        b.upgradeSchemaID,
		CreatedAt: chrono.TimeNow(),
		Balance:   b.wallet.Balance,
		Plan:      b.plan,
	}, nil
}

func (b *OrderBuilder) ProratedOrdersSchema() []ProratedOrderSchema {
	orders := make([]ProratedOrderSchema, 0)

	for _, v := range b.wallet.Source {
		s := ProratedOrderSchema{
			ProratedOrder: v,
			CreatedUTC:    chrono.TimeNow(),
			ConsumedUTC:   chrono.Time{},
			UpgradeID:     b.upgradeSchemaID,
		}

		orders = append(orders, s)
	}

	return orders
}

func (b *OrderBuilder) ClientApp() OrderClient {
	return OrderClient{
		OrderID:   b.orderID,
		ClientApp: b.client,
	}
}

func (b *OrderBuilder) AliAppPayParams() alipay.AliPayTradeAppPay {
	return alipay.AliPayTradeAppPay{
		TradePay: alipay.TradePay{
			NotifyURL:   b.getWebHookURL(),
			Subject:     b.plan.GetTitle(b.kind),
			OutTradeNo:  b.orderID,
			TotalAmount: b.charge.AliPrice(b.sandbox),
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
			Subject:     b.plan.GetTitle(b.kind),
			OutTradeNo:  b.orderID,
			TotalAmount: b.charge.AliPrice(b.sandbox),
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
			Subject:     b.plan.GetTitle(b.kind),
			OutTradeNo:  b.orderID,
			TotalAmount: b.charge.AliPrice(b.sandbox),
			ProductCode: ali.ProductCodeWeb.String(),
			GoodsType:   "0",
		},
	}
}

func (b *OrderBuilder) WxpayParams() wxpay.Params {
	p := make(wxpay.Params)
	p.SetString("body", b.plan.GetTitle(b.kind))
	p.SetString("out_trade_no", b.orderID)
	p.SetInt64("total_fee", b.charge.AmountInCent(b.sandbox))
	p.SetString("spbill_create_ip", b.client.UserIP.String)
	p.SetString("notify_url", b.getWebHookURL())
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
