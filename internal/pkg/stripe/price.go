package stripe

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"strconv"
	"time"
)

// PriceMetadata parsed the fields defined in stripe price's medata field.
// Those are customer-defined key-value pairs.
type PriceMetadata struct {
	Introductory bool            `json:"introductory"` // Is this an introductory price? If false, it is treated as regular price.
	PeriodCount  dt.YearMonthDay `json:"periodCount"`
	Tier         enum.Tier       `json:"tier"`     // The tier of this price.
	StartUTC     null.String     `json:"startUtc"` // Start time if Introductory is true; otherwise omit.
	EndUTC       null.String     `json:"endUtc"`   // End time if Introductory is true; otherwise omit.
}

// NewPriceMeta converts stripe price metadata to a struct.
// - tier: standard | premium
// - years: number
// - months: number
// - days: number
// - introductory: boolean
// - start_utc?: string
// - end_utc?: string
func NewPriceMeta(m map[string]string) PriceMetadata {
	tier, _ := enum.ParseTier(m["tier"])
	years, _ := strconv.Atoi(m["years"])
	months, _ := strconv.Atoi(m["months"])
	days, _ := strconv.Atoi(m["days"])
	isIntro, _ := strconv.ParseBool(m["introductory"])
	start, _ := m["start_utc"]
	end, _ := m["end_utc"]

	return PriceMetadata{
		Tier: tier,
		PeriodCount: dt.YearMonthDay{
			Years:  int64(years),
			Months: int64(months),
			Days:   int64(days),
		},
		Introductory: isIntro,
		StartUTC:     null.NewString(start, start != ""),
		EndUTC:       null.NewString(end, end != ""),
	}
}

func PriceMetaParams(p price.Price) map[string]string {

	start := ""
	end := ""
	if !p.StartUTC.IsZero() {
		start = p.StartUTC.In(time.UTC).Format(time.RFC3339)
	}

	if !p.EndUTC.IsZero() {
		end = p.EndUTC.In(time.UTC).Format(time.RFC3339)
	}

	return map[string]string{
		"tier":         p.Tier.String(),
		"years":        strconv.FormatInt(p.PeriodCount.Years, 10),
		"months":       strconv.FormatInt(p.PeriodCount.Months, 10),
		"days":         strconv.FormatInt(p.PeriodCount.Days, 10),
		"introductory": strconv.FormatBool(p.Kind == price.KindOneTime),
		"start_utc":    start,
		"end_utc":      end,
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
	ID         string           `json:"id"`
	Active     bool             `json:"active"`
	Created    int64            `json:"created"`
	Currency   stripe.Currency  `json:"currency"`
	LiveMode   bool             `json:"liveMode"`
	Metadata   PriceMetadata    `json:"metadata"` // Deprecated
	Nickname   string           `json:"nickname"`
	Product    string           `json:"product"`
	Recurring  PriceRecurring   `json:"recurring"`
	Type       stripe.PriceType `json:"type"`
	UnitAmount int64            `json:"unitAmount"`
	PriceMetadata
}

func NewPrice(p *stripe.Price) Price {
	return Price{
		Active:        p.Active,
		Created:       p.Created,
		Currency:      p.Currency,
		ID:            p.ID,
		LiveMode:      p.Livemode,
		Nickname:      p.Nickname,
		Product:       p.Product.ID,
		Recurring:     NewPriceRecurring(p.Recurring),
		Type:          p.Type,
		UnitAmount:    p.UnitAmount,
		PriceMetadata: NewPriceMeta(p.Metadata),
	}
}

func (p Price) IsZero() bool {
	return p.ID == ""
}

// Edition deduces the edition of stripe price since that
// information might be missing.
func (p Price) Edition() price.Edition {

	if p.Type == stripe.PriceTypeRecurring {
		cycle, err := enum.ParseCycle(string(p.Recurring.Interval))
		if err == nil {
			return price.Edition{
				Tier:  p.Tier,
				Cycle: cycle,
			}
		}
	}

	return price.Edition{
		Tier:  p.Tier,
		Cycle: p.PeriodCount.EqCycle(),
	}
}
