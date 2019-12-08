package subscription

import "errors"

var ErrLinkToMultipleFTC = errors.New("one apple subscription cannot be linked to multiple FTC account")

var ErrTargetLinkedToOtherIAP = errors.New("target ftc account is already linked to another apple subscription")

var ErrHasValidNonIAPMember = errors.New("target ftc account already has a valid membership")

var (
	// Cannot deduce subscription kind.
	ErrRenewalForbidden = errors.New("current membership expiration date exceeds max renewal period")
	ErrSubsUsageUnclear = errors.New("unknown subscription usage")
	ErrPlanRequired     = errors.New("membership is not set")

	// Reasons to deny upgrade.
	ErrUpgradeInvalid            = errors.New("upgrade could only be performed on an existing non-premium membership purchased via alipay or wxpay")
	ErrBalanceCannotCoverUpgrade = errors.New("current balance is not enough to cover upgrading cost")
)
