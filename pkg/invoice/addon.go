package invoice

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/collection"
	"time"
)

func ConsumeAddOns(addOns []Invoice) []Invoice {
	now := time.Now()

	startTime := now

	invoices := make([]Invoice, 0)
	for _, v := range addOns {
		consumed := v.SetPeriod(startTime)
		startTime = consumed.EndUTC.Time
		invoices = append(invoices, consumed)
	}

	return invoices
}

// The following are deprecated implementations.
func groupInvoices(addOns []Invoice) map[enum.Tier][]Invoice {
	g := make(map[enum.Tier][]Invoice)

	for _, v := range addOns {
		g[v.Tier] = append(g[v.Tier], v)
	}

	return g
}

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
	grouped := groupInvoices(addOns)

	var sums = make(map[enum.Tier]AddOnSum)

	for k, v := range grouped {
		if len(v) > 0 {
			sums[k] = NewAddOnSum(v)
		}
	}

	return sums
}
