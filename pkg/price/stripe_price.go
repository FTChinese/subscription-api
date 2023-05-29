package price

import (
	"fmt"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/conv"
	"github.com/stripe/stripe-go/v72"
)

type StripePrice struct {
	IsFromStripe   bool               `json:"-"`
	ID             string             `json:"id" db:"id"`
	Active         bool               `json:"active" db:"active"`
	Currency       stripe.Currency    `json:"currency" db:"currency"`
	IsIntroductory bool               `json:"isIntroductory" db:"is_intro"` // Deprecated
	Kind           Kind               `json:"kind" db:"kind"`
	LiveMode       bool               `json:"liveMode" db:"live_mode"`
	Nickname       string             `json:"nickname" db:"nickname"`
	OnPaywall      bool               `json:"onPaywall" db:"on_paywall"`
	ProductID      string             `json:"productId" db:"product_id"`
	PeriodCount    ColumnYearMonthDay `json:"periodCount" db:"period_count"`
	Tier           enum.Tier          `json:"tier" db:"tier"` // The tier of this price.
	UnitAmount     int64              `json:"unitAmount" db:"unit_amount"`
	StartUTC       chrono.Time        `json:"startUtc" db:"start_utc"` // Start time if Introductory is true; otherwise omit.
	EndUTC         chrono.Time        `json:"endUtc" db:"end_utc"`     // End time if Introductory is true; otherwise omit.
	Created        int64              `json:"created" db:"created"`

	Metadata  StripePriceMeta      `json:"metadata"`  // Deprecated
	Product   string               `json:"product"`   // Deprecated. Use ProductID
	Recurring StripePriceRecurring `json:"recurring"` // Deprecated
	Type      stripe.PriceType     `json:"type"`      // Deprecated. Use Kind.
}

func NewStripePrice(p *stripe.Price) StripePrice {

	meta := ParseStripePriceMeta(p.Metadata).
		WithRecurring(p.Recurring)

	return StripePrice{
		IsFromStripe:   true,
		ID:             p.ID,
		Active:         p.Active,
		Currency:       p.Currency,
		IsIntroductory: meta.Introductory,
		Kind:           Kind(p.Type),
		LiveMode:       p.Livemode,
		Nickname:       p.Nickname,
		ProductID:      p.Product.ID,
		PeriodCount:    ColumnYearMonthDay{meta.PeriodCount},
		Tier:           meta.Tier,
		UnitAmount:     p.UnitAmount,
		StartUTC:       meta.StartUTC,
		EndUTC:         meta.EndUTC,
		Created:        p.Created,

		Metadata:  meta,
		Product:   p.Product.ID,
		Recurring: NewPriceRecurring(p.Recurring),
		Type:      p.Type,
	}
}

func (p StripePrice) ActiveID() conv.MD5Sum {
	f := p.uniqueFeatures()
	return conv.NewMD5Sum(f)
}

func (p StripePrice) uniqueFeatures() string {
	e := p.Edition()

	return fmt.Sprintf("stripe.%s.%s.%s.%s", e.TierString(), e.CycleString(), p.Kind, conv.LiveMode(p.LiveMode))
}

func (p StripePrice) ActiveEntry() ActivePrice {
	return ActivePrice{
		ID:         p.ActiveID().ToHexBin(),
		Source:     PriceSourceStripe,
		ProductID:  p.ProductID,
		PriceID:    p.ID,
		UpdatedUTC: chrono.TimeUTCNow(),
	}
}

func (p StripePrice) IsZero() bool {
	return p.ID == ""
}

func (p StripePrice) IsIntro() bool {
	return p.Kind == KindOneTime
}

// Edition deduces the edition of stripe price since that
// information might be missing.
func (p StripePrice) Edition() Edition {
	if p.Kind == KindOneTime {
		return Edition{
			Tier:  p.Tier,
			Cycle: enum.CycleNull,
		}
	}

	return Edition{
		Tier:  p.Tier,
		Cycle: p.PeriodCount.EqCycle(),
	}
}

// StripePriceRecurring is the equivalence of stripe.PriceRecurring.
// Deprecated
type StripePriceRecurring struct {
	Interval      stripe.PriceRecurringInterval  `json:"interval"`
	IntervalCount int64                          `json:"intervalCount"`
	UsageType     stripe.PriceRecurringUsageType `json:"usageType"`
}

// NewPriceRecurring converts stripe recurring.
// Deprecated.
func NewPriceRecurring(r *stripe.PriceRecurring) StripePriceRecurring {
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
