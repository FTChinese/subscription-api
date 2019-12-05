package subscription

import "errors"

var ErrLinkToMultipleFTC = errors.New("one apple subscription cannot be linked to multiple FTC account")

var ErrTargetLinkedToOtherIAP = errors.New("target ftc account is already linked to another apple subscription")

var ErrHasValidNonIAPMember = errors.New("target ftc account already has a valid membership")
