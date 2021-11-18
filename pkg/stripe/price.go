package stripe

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"strconv"
)

// PriceMetadata parsed the fields defined in stripe price's medata field.
// Those are customer-defined key-value pairs.
type PriceMetadata struct {
	Tier         enum.Tier   `json:"tier"`         // The tier of this price.
	PeriodDays   int64       `json:"periodDays"`   // The days this price purchased.
	Introductory bool        `json:"introductory"` // Is this an introductory price? If false, it is treated as regular price.
	StartUTC     null.String `json:"startUtc"`     // Start time if Introductory is true; otherwise omit.
	EndUTC       null.String `json:"endUtc"`       // End time if Introductory is true; otherwise omit.
}

func NewStripePriceMeta(m map[string]string) PriceMetadata {
	t, _ := enum.ParseTier(m["tier"])
	d, _ := strconv.Atoi(m["period_days"])
	isIntro, _ := strconv.ParseBool(m["introductory"])
	start, _ := m["start_utc"]
	end, _ := m["end_utc"]

	return PriceMetadata{
		Tier:         t,
		PeriodDays:   int64(d),
		Introductory: isIntro,
		StartUTC:     null.NewString(start, start != ""),
		EndUTC:       null.NewString(end, end != ""),
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

// Edition deduces the edition of stripe price since that
// information might be missing.
func (p Price) Edition() price.Edition {
	se, err := PriceEditionStore.FindByID(p.ID)
	if err == nil {
		return se.Edition
	}

	if p.Type == stripe.PriceTypeRecurring {
		cycle, err := enum.ParseCycle(string(p.Recurring.Interval))
		if err == nil {
			return price.Edition{
				Tier:  p.Metadata.Tier,
				Cycle: cycle,
			}
		}
	}

	days := p.Metadata.PeriodDays

	if days >= 365 {
		return price.Edition{
			Tier:  p.Metadata.Tier,
			Cycle: enum.CycleYear,
		}
	}

	if days >= 30 && days <= 366 {
		return price.Edition{
			Tier:  p.Metadata.Tier,
			Cycle: enum.CycleMonth,
		}
	}

	return price.Edition{}
}
