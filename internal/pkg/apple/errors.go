package apple

import (
	"errors"
	"github.com/FTChinese/go-rest/render"
)

var (
	ErrAlreadyLinked = errors.New("iap already linked to the ftc account")
	// This error is a possible cheating.
	ErrIAPAlreadyLinked = errors.New("iap subscription is already claimed by another ftc account")
	// This error might be caused by one ftc account linking to multiple IAP.
	ErrFtcAlreadyLinked = errors.New("ftc account is already linked to a iap subscription")
	// Ftc account has a non-expired membership that comes from other payment method, thus we cannot override it.
	ErrFtcMemberValid    = errors.New("ftc account has valid membership purchased via non-apple method")
	ErrIAPAlreadyExpired = errors.New("iap subscription expired before ftc expires")
)

func ConvertLinkErr(err error) (*render.ValidationError, bool) {
	switch err {
	// Multiple FTC accounts linking to single IAP.
	case ErrIAPAlreadyLinked:
		return &render.ValidationError{
			Message: "Apple subscription is already claimed by another ftc account.",
			Field:   "originalTxId",
			Code:    "linked_to_other_ftc",
		}, true

	// Single FTC account linking to multiple IAP
	case ErrFtcAlreadyLinked:
		return &render.ValidationError{
			Message: "FTC account is already linked to another Apple subscription",
			Field:   "ftcId",
			Code:    "linked_to_other_iap",
		}, true

	case ErrFtcMemberValid:
		return &render.ValidationError{
			Message: "FTC account already has a valid membership via non-Apple channel",
			Field:   "ftcId",
			Code:    "has_valid_non_iap",
		}, true

	case ErrIAPAlreadyExpired:
		return &render.ValidationError{
			Message: "You are not allowed to link to an already expired IAP subscription",
			Field:   "originalTxId",
			Code:    "already_expired",
		}, true

	default:
		return nil, false
	}
}
