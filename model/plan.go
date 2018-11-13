package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"

	"gitlab.com/ftchinese/subscription-api/util"
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

// CreateOrderID creates the order number based on the plan selected.
func CreateOrderID(p Plan) string {
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between [100, 999)
	rn := 100 + rand.Intn(999-100+1)

	return fmt.Sprintf("FT%03d%d%d", p.ID, rn, time.Now().Unix())
}

// Schedule contains discount plans and duration.
// Start and end are all formatted to ISO8601 in UTC: 2006-01-02T15:04:05Z
type Schedule struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Start     string          `json:"startAt"`
	End       string          `json:"endAt"`
	Plans     map[string]Plan `json:"plans"`
	CreatedAt string          `json:"createdAt"`
	createdBy string
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

// RetrieveSchedule finds a lastest discount schedule whose end time is still after now.
func (env Env) RetrieveSchedule() (Schedule, error) {
	query := `
	SELECT
		id AS id,
		name AS name, 
		start_utc AS start,
		end_utc AS end,
		plans AS plans,
		created_utc AS createdUtc,
		created_by AS createdBy
	FROM premium.discount_schedule
	WHERE end_utc >= UTC_TIMESTAMP() 
	ORDER BY created_utc DESC
	LIMIT 1`

	var s Schedule
	var plans string
	var start string
	var end string

	err := env.DB.QueryRow(query).Scan(
		&s.ID,
		&s.Name,
		&start,
		&end,
		&plans,
		&s.CreatedAt,
		&s.createdBy,
	)

	if err != nil {
		return s, err
	}

	if err := json.Unmarshal([]byte(plans), &s.Plans); err != nil {
		return s, err
	}

	s.Start = util.ISO8601UTC.FromDatetime(start, nil)
	s.End = util.ISO8601UTC.FromDatetime(end, nil)
	s.CreatedAt = util.ISO8601UTC.FromDatetime(s.CreatedAt, nil)

	// Cache it.
	env.cacheSchedule(s)

	return s, nil
}

func (env Env) cacheSchedule(s Schedule) {
	env.Cache.Set(keySchedule, s, cache.NoExpiration)
}

// ScheduleFromCache gets schedule from cache.
func (env Env) ScheduleFromCache() (Schedule, bool) {
	if x, found := env.Cache.Get(keySchedule); found {
		sch, ok := x.(Schedule)

		if ok {
			logger.WithField("location", "ScheduleFromCache").Infof("Cached schedule found %+v", sch)
			return sch, true
		}

		return Schedule{}, false
	}

	return Schedule{}, false
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
