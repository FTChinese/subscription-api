package paywall

import "errors"

var (
	ErrTierMismatched     = errors.New("the tier of upgrading order is not used for upgrading")
	ErrDuplicateUpgrading = errors.New("duplicate upgrading")
	ErrNoUpgradingTarget  = errors.New("membership to upgrade not found")
)
