package invoice

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/collection"
)

// The following are deprecated implementations.
type AddOnSum struct {
	IDs    collection.StringSet
	Years  int
	Months int
	Days   int
	Latest Invoice
}

func NewAddOnSum(addOns []Invoice) AddOnSum {
	if len(addOns) == 0 {
		return AddOnSum{}
	}

	sum := AddOnSum{
		IDs:    make(collection.StringSet),
		Latest: addOns[len(addOns)-1], // The addon array should already be sorted in descending order by CreatedUTC field.
	}

	for _, v := range addOns {
		// Collect IDs so that we could mark those rows in DB as consumed.
		sum.IDs[v.ID] = nil

		sum.Years += int(v.Years)
		sum.Months += int(v.Months)
		sum.Days += int(v.Days)
	}

	return sum
}

func GroupAndReduce(addOns []Invoice) map[enum.Tier]AddOnSum {
	grouped := groupAddOns(addOns)

	var sums = make(map[enum.Tier]AddOnSum)

	for k, v := range grouped {
		if len(v) > 0 {
			sums[k] = NewAddOnSum(v)
		}
	}

	return sums
}

func groupAddOns(addOns []Invoice) AddOnGroup {
	g := make(map[enum.Tier][]Invoice)

	for _, v := range addOns {
		if v.OrderKind != enum.OrderKindAddOn {
			continue
		}
		g[v.Tier] = append(g[v.Tier], v)
	}

	return g
}
