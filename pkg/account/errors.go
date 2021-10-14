package account

import "errors"

var (
	ErrMobileTakenByOther = errors.New("the mobile phone is already taken by another account")
)
