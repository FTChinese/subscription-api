package subs

import "errors"

// Errors when deducing a SubsKind from current membership and the new plan to subscribe.
var (
	ErrRenewalForbidden   = errors.New("current membership expiration date exceeds max renewal period")
	ErrDowngradeForbidden = errors.New("downgrading membership is not allowed")
	ErrUnknownSubsKind    = errors.New("unknown subscription usage")
	// In case the plan is not set in OrderBuilder
	ErrInvalidPlan = errors.New("no pricing plan is provided")
)

var (
	// Reasons to deny upgrade.
	ErrUpgradeInvalid            = errors.New("upgrade could only be performed on an existing non-premium membership purchased via alipay or wxpay")
	ErrBalanceCannotCoverUpgrade = errors.New("current balance is not enough to cover upgrading cost")

	// For subscription via stripe
	ErrNonStripeValidSub = errors.New("you already have a valid subscription purchased via non-stripe method")
	ErrActiveStripeSub   = errors.New("you already have an active subscription purchased via stripe")
	ErrUnknownSubState   = errors.New("your subscription status cannot be determined")

	// For upgrading via stripe.
	ErrInvalidStripeSub = errors.New("you must have a valid stripe subscription to upgrade")
)
