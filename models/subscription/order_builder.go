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
	"time"
)

type OrderBuilder struct {
	// Required fields.
	memberID reader.MemberID
	plan     plan.Plan
	live     bool // Use live webhool or sandbox.

	membership Membership // User's current membership. Needs querying DB.
	wallet     Wallet     // Only required if kind == SubsKindUpgrade.

	// Optional fields.
	client          util.ClientApp      // Only requird when actually creating an order.
	method          enum.PayMethod      // Should be enum.PayMethodNull for free upgrade.
	wxAppID         null.String         // Only required for wechat pay.
	wxUnifiedParams wechat.UnifiedOrder // Only for wechat pay.

	// Calculated for previous fields set.
	kind            plan.SubsKind
	charge          plan.Charge
	duration        Duration
	orderID         string
	upgradeSchemaID null.String
	snapshotID      null.String // The ID of current membership's snapshot. Only present when member is not zero.

	isBuilt bool // A flag to ensure all fields are property initialized.
}

func NewOrderBuilder(id reader.MemberID) *OrderBuilder {
	return &OrderBuilder{
		memberID: id,
		duration: Duration{
			CycleCount: 1,
			ExtraDays:  1,
		},
		isBuilt: false,
	}
}

func (b *OrderBuilder) SetWxAppID(appID string) *OrderBuilder {
	b.wxAppID = null.StringFrom(appID)
	return b
}

func (b *OrderBuilder) GetReaderID() reader.MemberID {
	return b.memberID
}

func (b *OrderBuilder) SetEnvironment(live bool) *OrderBuilder {
	b.live = live
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

func (b *OrderBuilder) GetMembership() Membership {
	return b.membership
}

// buildSubsKind determines what kind of subscripiton
// it is deduced from user's current membership and
// chosen plan.
// Error: ErrRenewalForbidden, ErrSubsUsageUnclear, ErrPlanRequired.
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

// GetSubsKind returns the SubsKind, or error if it is
// cannot be deduced.
// See errors returned from buildSubsKind.
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

// CanBalanceCoverPlan checks if user's current balance
// is enough to cover chosen plan.
func (b *OrderBuilder) CanBalanceCoverPlan() bool {
	return b.wallet.GetBalance() >= b.plan.Amount
}

// Build calculates subscription kind, the amount to pay,
// billing cycles user purchased.
// See buildSubsKind form returned errors.
func (b *OrderBuilder) Build() error {

	if b.isBuilt {
		return nil
	}

	// Deduce subscription kind if not performed yet.
	if err := b.buildSubsKind(); err != nil {
		return err
	}

	// Wallet should exist only for upgrading order.
	if b.kind != plan.SubsKindUpgrade {
		b.wallet = Wallet{}
	}

	// A zero wallet still produces valid Duration value,
	// which default to 1 cycle plus 1 day.
	b.duration = b.wallet.ConvertBalance(b.plan)
	b.charge = plan.Charge{
		Amount:   b.plan.Price - b.wallet.GetBalance(),
		Currency: b.plan.Currency,
	}

	// If balance is larger than a plan's price,
	// charged amount should be zero.
	if b.charge.Amount < 0 {
		b.charge.Amount = 0
	}

	if b.kind == plan.SubsKindUpgrade && b.upgradeSchemaID.IsZero() {
		b.upgradeSchemaID = null.StringFrom(GenerateUpgradeID())
	}

	if !b.membership.IsZero() && b.snapshotID.IsZero() {
		b.snapshotID = null.StringFrom(GenerateSnapshotID())
	}

	b.isBuilt = true
	return nil
}

func (b *OrderBuilder) MustBuild() *OrderBuilder {
	if err := b.Build(); err != nil {
		panic(err)
	}

	return b
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

// Order creates a new order for
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
		BasePlan:         b.plan.BasePlan,
		Price:            b.plan.Price,
		Charge:           b.charge,
		Duration:         b.duration,
		Usage:            b.kind,
		PaymentMethod:    b.method,
		WxAppID:          b.wxAppID,
		UpgradeSchemaID:  b.upgradeSchemaID,
		StartDate:        chrono.Date{},
		EndDate:          chrono.Date{},
		CreatedAt:        chrono.TimeNow(),
		ConfirmedAt:      chrono.Time{},
		MemberSnapshotID: b.snapshotID, // Set when confirming order.
	}, nil
}

func (b *OrderBuilder) FreeUpgradeOrder() (Order, Membership, error) {
	o, err := b.Order()
	if err != nil {
		return Order{}, Membership{}, err
	}

	startTime := time.Now()
	endTime, err := o.getEndDate(startTime)
	if err != nil {
		return Order{}, Membership{}, err
	}
	o.StartDate = chrono.DateFrom(startTime)
	o.EndDate = chrono.DateFrom(endTime)
	o.ConfirmedAt = chrono.TimeNow()

	m := b.membership
	m.Tier = o.Tier
	m.Cycle = o.Cycle
	m.ExpireDate = o.EndDate

	return o, m, nil
}

func (b *OrderBuilder) MembershipSnapshot() MemberSnapshot {
	return MemberSnapshot{
		SnapshotID: b.snapshotID.String,
		Reason:     b.kind.SnapshotReason(),
		CreatedUTC: chrono.TimeNow(),
		Membership: b.membership,
	}
}

// UpgradeBalanceSchema converts wallet to save the total balance part of data.
func (b *OrderBuilder) UpgradeBalanceSchema() (UpgradeBalanceSchema, error) {

	if err := b.ensureBuilt(); err != nil {
		return UpgradeBalanceSchema{}, err
	}

	if b.kind != plan.SubsKindUpgrade {
		return UpgradeBalanceSchema{}, errors.New("not an upgrade subscription")
	}

	return UpgradeBalanceSchema{
		ID:         b.upgradeSchemaID.String,
		CreatedAt:  chrono.TimeNow(),
		Balance:    b.wallet.GetBalance(),
		PlanPrice:  b.plan.Price,
		PlanAmount: b.plan.Amount,
	}, nil
}

// ProratedOrdersSchema wallet to save what make up of a
// wallet's total balance.
func (b *OrderBuilder) ProratedOrdersSchema() []ProratedOrderSchema {
	orders := make([]ProratedOrderSchema, 0)

	for _, v := range b.wallet.Source {
		s := ProratedOrderSchema{
			ProratedOrder: v,
			CreatedUTC:    chrono.TimeNow(),
			ConsumedUTC:   chrono.Time{},
			UpgradeID:     b.upgradeSchemaID.String,
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
			TotalAmount: b.charge.AliPrice(b.live),
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
			TotalAmount: b.charge.AliPrice(b.live),
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
			TotalAmount: b.charge.AliPrice(b.live),
			ProductCode: ali.ProductCodeWeb.String(),
			GoodsType:   "0",
		},
	}
}

func (b *OrderBuilder) WxpayParams() wxpay.Params {
	p := make(wxpay.Params)
	p.SetString("body", b.plan.GetTitle(b.kind))
	p.SetString("out_trade_no", b.orderID)
	p.SetInt64("total_fee", b.charge.AmountInCent(b.live))
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
