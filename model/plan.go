package model

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// MemberTier represents membership tiers
type MemberTier string

// CN translates MemberTier into Chinese 标准会员 or 高级会员.
func (t MemberTier) CN() string {
	switch t {
	case TierStandard:
		return "标准会员"
	case TierPremium:
		return "高级会员"
	default:
		return ""
	}
}

// EN translates MemberTier into English.
func (t MemberTier) EN() string {
	switch t {
	case TierStandard:
		return "Standard"
	case TierPremium:
		return "Premium"
	default:
		return ""
	}
}

// BillingCycle is an enum of billing cycles.
type BillingCycle string

// CN translates BillingCycle into Chinese 年 or 月
func (c BillingCycle) CN() string {
	switch c {
	case Yearly:
		return "年"
	case Monthly:
		return "月"
	default:
		return ""
	}
}

// EN translates BillingCycle into Chinese 年 or 月
func (c BillingCycle) EN() string {
	switch c {
	case Yearly:
		return "Year"
	case Monthly:
		return "Month"
	default:
		return ""
	}
}

// PaymentMethod lists supported payment channels.
type PaymentMethod string

// CN translates PaymentMethod into Chinese text.
func (m PaymentMethod) CN() string {
	switch m {
	case Alipay:
		return "支付宝"
	case Wxpay:
		return "微信支付"
	case Stripe:
		return "Stripe"
	default:
		return ""
	}
}

// EN translates PaymentMethod into English text.
func (m PaymentMethod) EN() string {
	switch m {
	case Alipay:
		return "Ali Pay"
	case Wxpay:
		return "Wechat Pay"
	case Stripe:
		return "Stripe"
	default:
		return ""
	}
}

const (
	// TierInvalid is a placeholder
	TierInvalid MemberTier = ""
	// TierStandard is the standard tier
	TierStandard MemberTier = "standard"
	// TierPremium is the premium tier
	TierPremium MemberTier = "premium"
	// CycleInvalid is a placeholder
	CycleInvalid BillingCycle = ""
	// Yearly bills every year
	Yearly BillingCycle = "year"
	// Monthly bills every month
	Monthly BillingCycle = "month"
	// Alipay supports taobao payment
	Alipay PaymentMethod = "alipay"
	// Wxpay supports wechat payment
	Wxpay PaymentMethod = "tenpay"
	// Stripe supports pay by stripe
	Stripe PaymentMethod = "stripe"
)

// NewTier returns a MemberTier.
// `key` either `standard` or `premium`
func NewTier(key string) (MemberTier, error) {
	switch key {
	case "standard":
		return TierStandard, nil

	case "premium":
		return TierPremium, nil

	default:
		return MemberTier(""), errors.New("Only standard and premium tier allowed")
	}
}

// NewCycle returns a new BillingCycle.
// `key` is either `year` or `month`.
func NewCycle(key string) (BillingCycle, error) {
	switch key {
	case "year":
		return Yearly, nil
	case "month":
		return Monthly, nil
	default:
		return CycleInvalid, errors.New("cycle must either be year or month")
	}
}

// Plan represents a subscription plan
type Plan struct {
	Tier        MemberTier   `json:"tier"`
	Cycle       BillingCycle `json:"cycle"`
	Price       float64      `json:"price"`
	ID          int          `json:"id"` // 10 for standard and 100 for premium
	Description string       `json:"description"`
}

// GetPriceCent calculates price in cent to be used for Wechat pay.
func (p Plan) GetPriceCent() int64 {
	return int64(p.Price * 100)
}

// GetPriceString formats price for alipay
func (p Plan) GetPriceString() string {
	return strconv.FormatFloat(p.Price, 'f', 2, 32)
}

// CreateOrder generates a new subscription order based on the plan chosen.
func (p Plan) CreateOrder(userID string, method PaymentMethod) Subscription {
	return Subscription{
		OrderID:       CreateOrderID(p),
		TierToBuy:     p.Tier,
		BillingCycle:  p.Cycle,
		Price:         p.Price,
		TotalAmount:   p.Price,
		PaymentMethod: method,
		UserID:        userID,
	}
}

// CreateOrderID creates the order number based on the plan selected.
func CreateOrderID(p Plan) string {
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between [100, 999)
	rn := 100 + rand.Intn(999-100+1)

	return fmt.Sprintf("FT%03d%d%d", p.ID, rn, time.Now().Unix())
}

// DefaultPlans is the default subscription. No discount.
var DefaultPlans = map[string]Plan{
	"standard_year": Plan{
		Tier:        TierStandard,
		Cycle:       Yearly,
		Price:       198.00,
		ID:          10,
		Description: "FT中文网 - 年度标准会员",
	},
	"standard_month": Plan{
		Tier:        TierStandard,
		Cycle:       Monthly,
		Price:       28.00,
		ID:          5,
		Description: "FT中文网 - 月度标准会员",
	},
	"premium_year": Plan{
		Tier:        TierPremium,
		Cycle:       Yearly,
		Price:       1998.00,
		ID:          100,
		Description: "FT中文网 - 高端会员",
	},
}

// DiscountSchedule and their duration.
var DiscountSchedule = Schedule{
	Start: "2018-10-01T16:00:00Z",
	End:   "2018-10-31T16:00:00Z",
	Plans: map[string]Plan{
		"standard_year": Plan{
			Tier:        TierStandard,
			Cycle:       Yearly,
			Price:       0.01,
			ID:          10,
			Description: "FT中文网 - 年度标准会员",
		},
		"standard_month": Plan{
			Tier:        TierStandard,
			Cycle:       Monthly,
			Price:       0.01,
			ID:          5,
			Description: "FT中文网 - 月度标准会员",
		},
		"premium_year": Plan{
			Tier:        TierPremium,
			Cycle:       Yearly,
			Price:       0.01,
			ID:          100,
			Description: "FT中文网 - 高端会员",
		},
	},
}

// GetCurrentPlans get default plans or discount plans depending on current time.
func (env Env) GetCurrentPlans() map[string]Plan {

	sch, found := env.ScheduleFromCache()

	// If no cache is found, use default ones.
	if !found {
		logger.WithField("location", "GetCurrentPlans").Info("Cached discount schedule not found. Use defualt plans")
		return DefaultPlans
	}

	// If cache is found, compare time
	now := time.Now()
	start := parseISO8601(sch.Start)
	end := parseISO8601(sch.End)

	if now.Before(start) || now.After(end) {
		logger.WithField("location", "GetCurrentPlans").Info("Cached plans duration not effective. Use default ones")
		return DefaultPlans
	}

	logger.WithField("location", "GetCurrentPlans").Info("Using discount plans")

	return sch.Plans
}

// FindPlan picks a Plan instance depending
// on the member tier and billing cycle.
// Returns error if member tier or billing cycyle are not the predefined ones.
func (env Env) FindPlan(tier MemberTier, cycle BillingCycle) (Plan, error) {
	key := string(tier) + "_" + string(cycle)

	plans := env.GetCurrentPlans()
	p, ok := plans[key]

	if !ok {
		return p, errors.New("subscription plan not found")
	}

	return p, nil
}

func parseISO8601(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)

	if err != nil {
		return time.Now()
	}

	return t
}
