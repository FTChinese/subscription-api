package reader

type SubsKind int

const (
	SubsKindZero SubsKind = iota
	SubsKindNew
	SubsKindOneTimeToSub // Same as new, with valid remaining membership period.
	SubsKindSwitchCycle  // Switching subscription billing cycle, e.g., from month to year.
	SubsKindUpgrade      // Switching subscription tier, e.g., from standard to premium.
	SubsKindRefresh
)

var subsKindsLocalized = []string{
	"",
	"create new subscription",
	"switch from one-time purchase to subscription mode",
	"switch billing cycle",
	"upgrade to premium",
}

func (k SubsKind) IsNewSubs() bool {
	return k == SubsKindNew || k == SubsKindOneTimeToSub
}

func (k SubsKind) IsUpdating() bool {
	return k == SubsKindUpgrade || k == SubsKindSwitchCycle
}

func (k SubsKind) Localize() string {
	if k >= SubsKindZero && k <= SubsKindUpgrade {
		return subsKindsLocalized[k]
	}

	return ""
}
