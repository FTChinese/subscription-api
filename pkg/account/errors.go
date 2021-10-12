package account

import "errors"

var (
	ErrMobileTaken         = errors.New("the mobile phone is already taken by another account")
	ErrMobileSet           = errors.New("the mobile already set on this account")
	ErrAccountHasMobileSet = errors.New("this account already linked to another mobile")
)
