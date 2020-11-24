package apple

import "errors"

var (
	// This error is a possible cheating.
	ErrIAPAlreadyLinked = errors.New("iap subscription is already taken by another ftc account")
	// This error might be caused by one ftc account linking to multiple IAP.
	ErrFtcAlreadyLinked = errors.New("ftc account is already linked to an iap subscription")
	// Ftc account has a non-expired membership that comes from other payment method, thus we cannot override it.
	ErrFtcMemberValid    = errors.New("ftc account has non-expired, non-iap membership")
	ErrIAPAlreadyExpired = errors.New("iap subscription already expired")
)
