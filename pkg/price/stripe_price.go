package price

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/stripe/stripe-go/v72"
	"strconv"
)

type StripePriceMeta struct {
	Tier         enum.Tier `json:"tier"`
	PeriodDays   int64     `json:"periodDays"`
	Introductory bool      `json:"introductory"`
}

func NewStripePriceMeta(m map[string]string) StripePriceMeta {
	t, _ := enum.ParseTier(m["tier"])
	d, _ := strconv.Atoi(m["period_days"])
	ok, _ := strconv.ParseBool(m["introductory"])
	return StripePriceMeta{
		Tier:         t,
		PeriodDays:   d,
		Introductory: ok,
	}
}

type StripePriceRecurring struct {
	Interval      stripe.PriceRecurringInterval  `json:"interval"`
	IntervalCount int64                          `json:"intervalCount"`
	UsageType     stripe.PriceRecurringUsageType `json:"usageType"`
}

func NewStripePriceRecurring(r *stripe.PriceRecurring) StripePriceRecurring {
	if r == nil {
		return StripePriceRecurring{}
	}

	return StripePriceRecurring{
		Interval:      r.Interval,
		IntervalCount: r.IntervalCount,
		UsageType:     r.UsageType,
	}
}

func (r StripePriceRecurring) IsZero() bool {
	return r.Interval == "" && r.IntervalCount == 0 && r.UsageType == ""
}

type StripePrice struct {
	Active     bool                 `json:"active"`
	Created    int64                `json:"created"`
	Currency   stripe.Currency      `json:"currency"`
	ID         string               `json:"id"`
	LiveMode   bool                 `json:"liveMode"`
	Metadata   StripePriceMeta      `json:"metadata"`
	Nickname   string               `json:"nickname"`
	Product    string               `json:"product"`
	Recurring  StripePriceRecurring `json:"recurring"`
	Type       stripe.PriceType     `json:"type"`
	UnitAmount int64                `json:"unitAmount"`
}

func NewStripePrice(p *stripe.Price) StripePrice {
	return StripePrice{
		Active:     p.Active,
		Created:    p.Created,
		Currency:   p.Currency,
		ID:         p.ID,
		LiveMode:   p.Livemode,
		Metadata:   NewStripePriceMeta(p.Metadata),
		Nickname:   p.Nickname,
		Product:    p.Product.ID,
		Recurring:  NewStripePriceRecurring(p.Recurring),
		Type:       p.Type,
		UnitAmount: p.UnitAmount,
	}
}
