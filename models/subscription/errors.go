package subscription

import "errors"

var ErrValidFTCMember = errors.New("linking target has valid membership")

var ErrLinkToMultipleFTC = errors.New("one apple subscription cannot be linked to multiple FTC account")

var ErrLinkTargetAlreadyTaken = errors.New("target ftc account is already linked to another apple subscription")

var ErrLinkToExistingMember = errors.New("target ftc account already has a valid membership")
