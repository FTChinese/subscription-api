package cart

type SubsKind int

const (
	SubsKindNull SubsKind = iota
	SubsKindNew
	SubsKindOneTimeToStripe // Same as new, with valid remaining membership period.
	SubsKindSwitchCycle     // Switching subscription billing cycle, e.g., from month to year.
	SubsKindUpgrade         // Switching subscription tier, e.g., from standard to premium.
)

var subsKindsLocalized = []string{
	"",
	"create new subscription",
	"switch from one-time purchase to subscription mode",
	"switch billing cycle",
	"upgrade to premium",
}

func (k SubsKind) Localize() string {
	if k >= SubsKindNull && k <= SubsKindUpgrade {
		return subsKindsLocalized[k]
	}

	return ""
}
