package reader

import "errors"

// Errors when deducing a SubsKind from current membership and the new plan to subscribe.
var (
	ErrRenewalForbidden   = errors.New("current membership expiration date exceeds max renewal period")
	ErrDowngradeForbidden = errors.New("downgrading membership is not allowed")
	ErrUnknownSubsKind    = errors.New("unknown subscription usage")
)

// For subscription via stripe
var (
	ErrNonStripeValidSub = errors.New("you already have a valid subscription purchased via non-stripe method")
	ErrActiveStripeSub   = errors.New("you already have an active subscription purchased via stripe")
	ErrUnknownSubState   = errors.New("your subscription status cannot be determined")
)
