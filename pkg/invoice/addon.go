package invoice

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
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

func (g AddOnGroup) ToAddOn() addon.AddOn {
	return addon.AddOn{
		Standard: reduceInvoices(g[enum.TierStandard]),
		Premium:  reduceInvoices(g[enum.TierPremium]),
	}
}

// Consume addon invoices of the specified tier.
// The last item in the result slice should be used to update
// membership.
func (g AddOnGroup) Consume(tier enum.Tier, start time.Time) []Invoice {
	return consumeAddOn(g[tier], start)
}

// reduceInvoices calculate the total days of invoices of the same tier.
func reduceInvoices(invs []Invoice) int64 {
	var sum dt.YearMonthDay

	for _, v := range invs {
		sum = sum.Add(v.YearMonthDay)
	}

	return sum.TotalDays()
}

// consumeAddOn add the start and end time to a list of invoices,
// with each one's start time following the previous one's end time.
// The last invoice's end time should the the membership's expiration date.
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
