package cart

type SubsKind int

const (
	SubsKindNull SubsKind = iota
	SubsKindNew
	SubsKindOneTimeToStripe // Same as new, with valid remaining membership period.
	SubsKindSwitchCycle     // Switching subscription billing cycle, e.g., from month to year.
	SubsKindUpgrade         // Switching subscription tier, e.g., from standard to premium.
)
