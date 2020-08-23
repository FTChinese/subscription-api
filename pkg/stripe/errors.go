package stripe

import "errors"

// For upgrading via stripe.

var (
	ErrInvalidStripeSub = errors.New("you must have a valid stripe subscription to upgrade")
)
