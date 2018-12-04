package util

import "errors"

var (
	// ErrRenewalForbidden indicates a user is not allowed to renew membership
	ErrRenewalForbidden = errors.New("membership renewal forbidden")
	// ErrIncompatible is returned by Scan interface
	ErrIncompatible = errors.New("incompatible type to scan")
)
