package addon

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/collection"
)

// AddOnSum contains the total years, months, and days of user's add-on belong to the same tier.
type Sum struct {
	IDs    collection.StringSet
	Years  int
	Months int
	Days   int
	Latest AddOn // Use to fill the edition, payment method, and plan id fields of membership.
}

// GroupAddOns put add-ons into different groups by tier so that we won't mix
// different tiers when sum up reserved days.
func group(addOns []AddOn) map[enum.Tier][]AddOn {
	g := make(map[enum.Tier][]AddOn)

	for _, v := range addOns {
		g[v.Tier] = append(g[v.Tier], v)
	}

	return g
}

func reduce(addOns []AddOn) Sum {
	if len(addOns) == 0 {
		return Sum{}
	}

	sum := Sum{
		IDs:    make(collection.StringSet),
		Latest: addOns[0], // The addon array should already be sorted in descending order by CreatedUTC field.
	}

	for _, v := range addOns {
		// Collect IDs so that we could mark those rows in DB as consumed.
		sum.IDs[v.ID] = nil

		switch v.Cycle {
		case enum.CycleYear:
			sum.Years += int(v.CycleCount)

		case enum.CycleMonth:
			sum.Months += int(v.CycleCount)
		}

		sum.Days += int(v.DaysRemained)
	}

	return sum
}

func GroupAndReduce(addOns []AddOn) map[enum.Tier]Sum {
	grouped := group(addOns)

	var sums = make(map[enum.Tier]Sum)

	for k, v := range grouped {
		if len(v) > 0 {
			sums[k] = reduce(v)
		}
	}

	return sums
}
