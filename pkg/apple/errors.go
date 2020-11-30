package apple

import "errors"

var (
	ErrAlreadyLinked = errors.New("iap already linked to the ftc account")
	// This error is a possible cheating.
	ErrIAPAlreadyLinked = errors.New("iap subscription is already claimed by another ftc account")
	// This error might be caused by one ftc account linking to multiple IAP.
	ErrFtcAlreadyLinked = errors.New("ftc account is already linked to a iap subscription")
	// Ftc account has a non-expired membership that comes from other payment method, thus we cannot override it.
	ErrFtcMemberValid    = errors.New("ftc account has valid membership purchased via non-apple method")
	ErrIAPAlreadyExpired = errors.New("iap subscription expired before ftc expires")
)
