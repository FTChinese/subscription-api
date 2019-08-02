package util

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

// Shared errors.
var (
	ErrAlreadyExists = errors.New("already exists")

	ErrAlreadyConfirmed     = errors.New("subscription order is already confirmed")
	ErrAlreadyUpgraded      = errors.New("cannot upgrade: already a premium member")
	ErrMemberNotFound       = errors.New("requested user is not a member yet")
	ErrMemberExpired        = errors.New("your subscription is already expired")
	ErrValidStripeSwitching = errors.New("valid stripe user cannot subscribe via alipay or wxpay")

	ErrTierMismatched     = errors.New("the tier of upgrading order is not used for upgrading")
	ErrDuplicateUpgrading = errors.New("duplicate upgrading")
	ErrNoUpgradingTarget  = errors.New("membership to upgrade not found")
	ErrDowngrade          = errors.New("membership downgrading is not allowed")
	ErrBeyondRenewal      = errors.New("exceed maximum allowed membership duration")

	// For subscription via stripe
	ErrNonStripeValidSub = errors.New("you already have a valid subscription purchased via non-stripe method")
	ErrActiveStripeSub   = errors.New("you already have an active subscription purchased via stripe")
	ErrUnknownSubState   = errors.New("your subscription status cannot be determined")

	// For upgrading via stripe.
	ErrInvalidStripeSub = errors.New("you must have a valid stripe subscription to upgrade")
)

// IsAlreadyExists tests if SQL error is an duplicate error.
func IsAlreadyExists(err error) bool {
	if e, ok := err.(*mysql.MySQLError); ok && e.Number == 1062 {
		return true
	}

	if err == ErrAlreadyExists {
		return true
	}

	return false
}

var ()
