package apple

import "errors"

var (
	ErrUnlinkMismatchedFTC = errors.New("IAP subscription is not linked to the provided ftc id")
)
