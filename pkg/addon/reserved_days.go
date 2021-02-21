package addon

import "github.com/FTChinese/go-rest/enum"

// ReservedDays contains the subscription period that will become effective once current membership expired.
type ReservedDays struct {
	Standard int64 `json:"standardAddOn" db:"standard_addon"`
	Premium  int64 `json:"premiumAddOn" db:"premium_addon"`
}

func (d ReservedDays) Plus(other ReservedDays) ReservedDays {
	return ReservedDays{
		Standard: d.Standard + other.Standard,
		Premium:  d.Premium + other.Premium,
	}
}

func (d ReservedDays) Clear(tier enum.Tier) ReservedDays {
	switch tier {
	case enum.TierStandard:
		d.Standard = 0

	case enum.TierPremium:
		d.Premium = 0
	}

	return d
}
