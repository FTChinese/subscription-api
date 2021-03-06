package addon

import "github.com/FTChinese/go-rest/enum"

// AddOn contains the subscription period that will become effective once current membership expired.
type AddOn struct {
	Standard int64 `json:"standardAddOn" db:"standard_addon"`
	Premium  int64 `json:"premiumAddOn" db:"premium_addon"`
}

func (d AddOn) Plus(other AddOn) AddOn {
	return AddOn{
		Standard: d.Standard + other.Standard,
		Premium:  d.Premium + other.Premium,
	}
}

func (d AddOn) Clear(tier enum.Tier) AddOn {
	switch tier {
	case enum.TierStandard:
		d.Standard = 0

	case enum.TierPremium:
		d.Premium = 0
	}

	return d
}
