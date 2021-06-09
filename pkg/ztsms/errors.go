package ztsms

import "errors"

var (
	ErrMobileAlreadySet    = errors.New("mobile already set")
	ErrMobileAlreadyExists = errors.New("another mobile set to this account")
)
