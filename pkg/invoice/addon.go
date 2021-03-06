package invoice

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/reader"
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
func (g AddOnGroup) Consumable(start time.Time) ConsumedAddOns {
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

// ConsumedAddOns is a slice of add-on invoices that
// are consumed.
type ConsumedAddOns []Invoice

// ClaimedBy transfers the latest add-on invoice's end time
// to membership.
func (c ConsumedAddOns) ClaimedBy(m reader.Membership) (AddOnClaimed, error) {
	if len(c) == 0 {
		return AddOnClaimed{}, errors.New("no addon invoice found")
	}

	latest := c[len(c)-1]
	newM, err := latest.TransferAddOn(m)
	if err != nil {
		return AddOnClaimed{}, err
	}

	return AddOnClaimed{
		Invoices:   c,
		Membership: newM,
		Snapshot:   m.Snapshot(reader.FtcArchiver(enum.OrderKindAddOn)),
	}, nil
}

// AddOnClaimed contains the data that should be
// used to perform add-on transfer.
type AddOnClaimed struct {
	Invoices   []Invoice         // The invoices used to extend membership's expiration date. They will not be used next time.
	Membership reader.Membership // Update membership.
	Snapshot   reader.MemberSnapshot
}

func NewAddOnClaimed(addOns []Invoice, m reader.Membership) (AddOnClaimed, error) {
	return NewAddOnGroup(addOns).
		Consumable(dt.PickLater(time.Now(), m.ExpireDate.Time)).
		ClaimedBy(m)
}
