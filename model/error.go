package model

import "errors"

// Errors for controller to identify actions to take.
var (
	ErrAlreadyConfirmed = errors.New("subscription order is already confirmed")
	ErrAllowRetry       = errors.New("resend notification")
	ErrDenyRetry        = errors.New("no longer resend notification")
	ErrAlreadyUpgraded  = errors.New("cannot upgrade: already a premium member")
)
