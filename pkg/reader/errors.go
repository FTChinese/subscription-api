package reader

import "errors"

var (
	ErrAccountsAlreadyLinked = errors.New("accounts already linked")
)

// The following lists Stripe subscription errors
var (
	ErrTrialUpgradeForbidden = errors.New("upgrading in trialing period is not allowed")
	ErrAlreadyStripeSubs     = errors.New("already a stripe subscription")
	ErrAlreadyAppleSubs      = errors.New("already subscribed via apple")
	ErrAlreadyB2BSubs        = errors.New("already subscribed via B2B")
	ErrUnknownPaymentMethod  = errors.New("unknown payment for current subscription")
)
