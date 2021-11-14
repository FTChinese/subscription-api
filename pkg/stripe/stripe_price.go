package stripe

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/stripe/stripe-go/v72"
	"strconv"
)

type PriceMetadata struct {
	Tier         enum.Tier `json:"tier"`
	PeriodDays   int64     `json:"periodDays"`
	Introductory bool      `json:"introductory"`
}

func NewStripePriceMeta(m map[string]string) PriceMetadata {
	t, _ := enum.ParseTier(m["tier"])
	d, _ := strconv.Atoi(m["period_days"])
	ok, _ := strconv.ParseBool(m["introductory"])
	return PriceMetadata{
		Tier:         t,
		PeriodDays:   int64(d),
		Introductory: ok,
	}
}

type PriceRecurring struct {
	Interval      stripe.PriceRecurringInterval  `json:"interval"`
	IntervalCount int64                          `json:"intervalCount"`
	UsageType     stripe.PriceRecurringUsageType `json:"usageType"`
}

func NewPriceRecurring(r *stripe.PriceRecurring) PriceRecurring {
	if r == nil {
		return PriceRecurring{}
	}

	return PriceRecurring{
		Interval:      r.Interval,
		IntervalCount: r.IntervalCount,
		UsageType:     r.UsageType,
	}
}

func (r PriceRecurring) IsZero() bool {
	return r.Interval == "" && r.IntervalCount == 0 && r.UsageType == ""
}

type Price struct {
	Active     bool             `json:"active"`
	Created    int64            `json:"created"`
	Currency   stripe.Currency  `json:"currency"`
	ID         string           `json:"id"`
	LiveMode   bool             `json:"liveMode"`
	Metadata   PriceMetadata    `json:"metadata"`
	Nickname   string           `json:"nickname"`
	Product    string           `json:"product"`
	Recurring  PriceRecurring   `json:"recurring"`
	Type       stripe.PriceType `json:"type"`
	UnitAmount int64            `json:"unitAmount"`
}

func NewPrice(p *stripe.Price) Price {
	return Price{
		Active:     p.Active,
		Created:    p.Created,
		Currency:   p.Currency,
		ID:         p.ID,
		LiveMode:   p.Livemode,
		Metadata:   NewStripePriceMeta(p.Metadata),
		Nickname:   p.Nickname,
		Product:    p.Product.ID,
		Recurring:  NewPriceRecurring(p.Recurring),
		Type:       p.Type,
		UnitAmount: p.UnitAmount,
	}
}

func (p Price) IsZero() bool {
	return p.ID == ""
}
