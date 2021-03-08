package invoice

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"time"
)

// AddOnGroup contains a user's invoices grouped by tier.
type AddOnGroup map[enum.Tier][]Invoice

// NewAddOnGroup groups and filters a slice of invoices by tier.
// These invoices must have order kind addon.
//
// Usage:
// claimed, err := NewAddOnGroup(invoices).
//		Consumable(startTime).
//		ClaimedBy(currentMembership)
func NewAddOnGroup(inv []Invoice) AddOnGroup {
	g := make(map[enum.Tier][]Invoice)

	for _, v := range inv {
		if v.OrderKind != enum.OrderKindAddOn {
			continue
		}
		g[v.Tier] = append(g[v.Tier], v)
	}

	return g
}

// Consumable selects from the grouped invoices whose
// purchased period is set and can be transferred to
// membership. Premium invoices will be used if exists,
// then fallback to standard edition.
func (g AddOnGroup) Consumable(start time.Time) []Invoice {
	prmAddOns, ok := g[enum.TierPremium]
	if ok {
		return consumeAddOn(prmAddOns, start)
	}

	stdAddOns, ok := g[enum.TierStandard]
	if ok {
		return consumeAddOn(stdAddOns, start)
	}

	return []Invoice{}
}

func consumeAddOn(addOns []Invoice, start time.Time) []Invoice {
	now := chrono.TimeNow()

	invoices := make([]Invoice, 0)
	for _, v := range addOns {
		if v.IsConsumed() {
			continue
		}
		consumed := v.SetPeriod(start)
		consumed.ConsumedUTC = now
		start = consumed.EndUTC.Time
		invoices = append(invoices, consumed)
	}

	return invoices
}
