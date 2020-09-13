package reader

import "errors"

// For subscription via stripe
var (
	ErrNonStripeValidSub = errors.New("you already have a valid subscription purchased via non-stripe method")
	ErrActiveStripeSub   = errors.New("you already have an active subscription purchased via stripe")
	ErrUnknownSubState   = errors.New("your subscription status cannot be determined")
	ErrIAPFtcLinked      = errors.New("iap and ftc account already linked")
)
