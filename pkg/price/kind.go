package price

import "github.com/FTChinese/go-rest/enum"

// Kind represents the type of price.
type Kind string

const (
	KindRecurring Kind = "recurring"
	KindOneTime   Kind = "one_time"
)

// cycleStrOfKind produces a string of a price's recurring cycle.
// If price if of one_time, then the cycle is string null.
func cycleStrOfKind(k Kind, c enum.Cycle) string {
	if k == KindOneTime {
		return "null"
	}

	return c.String()
}
