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

// BillingCycle is an enum of billing cycles.
type BillingCycle string

// PaymentMethod lists supported payment channels.
type PaymentMethod string

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
	Tier        MemberTier
	Cycle       BillingCycle
	Price       float64
	ID          int // 10 for standard and 100 for premium
	Description string
}

// GetPriceCent calculates price in cent to be used for Wechat pay.
func (p Plan) GetPriceCent() int64 {
	return int64(p.Price * 100)
}

// GetPriceAli formats price for alipay
func (p Plan) GetPriceAli() string {
	return strconv.FormatFloat(p.Price, 'f', 2, 32)
}

// CreateOrderID creates the order number based on the plan selected.
func CreateOrderID(p Plan) string {
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between [100, 999)
	rn := 100 + rand.Intn(999-100+1)

	return fmt.Sprintf("FT%03d%d%d", p.ID, rn, time.Now().Unix())
}

// Discount contains discount plans and duration.
// Start and end are all formatted to ISO8601 in UTC: 2006-01-02T15:04:05Z
type Discount struct {
	Start string          `json:"startAt"`
	End   string          `json:"endAt"`
	Plans map[string]Plan `json:"plans"`
}

// DefaultPlans is the default subscription. No discount.
var DefaultPlans = map[string]Plan{
	"standard_year": Plan{
		Tier:        TierStandard,
		Cycle:       Yearly,
		Price:       198.00,
		ID:          10,
		Description: "FT中文网 - 标准会员",
	},
	"standard_month": Plan{
		Tier:        TierStandard,
		Cycle:       Monthly,
		Price:       28.00,
		ID:          5,
		Description: "FT中文网 - 标准会员",
	},
	"premium_year": Plan{
		Tier:        TierPremium,
		Cycle:       Yearly,
		Price:       1998.00,
		ID:          100,
		Description: "FT中文网 - 高端会员",
	},
}

// DiscountPlans and their duration.
var DiscountPlans = Discount{
	Start: "2018-10-01T16:00:00Z",
	End:   "2018-10-31T16:00:00Z",
	Plans: map[string]Plan{
		"standard_year": Plan{
			Tier:        TierStandard,
			Cycle:       Yearly,
			Price:       0.01,
			ID:          10,
			Description: "FT中文网 - 标准会员",
		},
		"standard_month": Plan{
			Tier:        TierStandard,
			Cycle:       Monthly,
			Price:       0.01,
			ID:          5,
			Description: "FT中文网 - 标准会员",
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
func GetCurrentPlans() map[string]Plan {
	now := time.Now()
	start := parseISO8601(DiscountPlans.Start)
	end := parseISO8601(DiscountPlans.End)

	if now.Before(start) || now.After(end) {
		return DefaultPlans
	}

	return DiscountPlans.Plans
}

// NewPlan creates a new Plan instance depending on the member tier and billing cycle chosen.
// Returns error if member tier or billing cycyle are not the predefined ones.
func NewPlan(tier MemberTier, cycle BillingCycle) (Plan, error) {
	key := string(tier) + "_" + string(cycle)

	plans := GetCurrentPlans()
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
