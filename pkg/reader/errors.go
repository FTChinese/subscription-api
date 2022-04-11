package reader

import (
	"errors"
	"github.com/FTChinese/go-rest/render"
)

var (
	ErrAccountsAlreadyLinked = errors.New("accounts already linked")
)

// The following lists errors contained in CheckoutIntent for stripe.
var (
	ErrTrialUpgradeForbidden = errors.New("upgrading in trialing period is not allowed")
	ErrAlreadyStripeSubs     = errors.New("already a stripe subscription")
	ErrAlreadyAppleSubs      = errors.New("already subscribed via apple")
	ErrAlreadyB2BSubs        = errors.New("already subscribed via B2B")
	ErrUnknownPaymentMethod  = errors.New("unknown payment for current subscription")
)

// Errors in CheckoutIntent for one-time purchase
var (
	ErrExceedingMaxRenewal   = errors.New("exceeding allowed max renewal period")
	ErrSubsUpgradeViaOneTime = errors.New("subscription mode cannot use one-time purchase to upgrade")
	ErrB2BUpgradeViaOneTime  = errors.New("corporate subscription cannot use retail payment to upgrade")
)

func ConvertIntentError(err error) error {
	switch err {
	case ErrAlreadyStripeSubs,
		ErrAlreadyAppleSubs,
		ErrAlreadyB2BSubs:
		return &render.ValidationError{
			Message: err.Error(),
			Field:   "membership",
			Code:    render.CodeAlreadyExists,
		}

	case ErrTrialUpgradeForbidden:
		return &render.ValidationError{
			Message: err.Error(),
			Field:   "trial_upgrade",
			Code:    render.CodeInvalid,
		}

	case ErrExceedingMaxRenewal:
		return &render.ValidationError{
			Message: err.Error(),
			Field:   "exceeding_max_renewal",
			Code:    render.CodeInvalid,
		}

	case ErrSubsUpgradeViaOneTime,
		ErrB2BUpgradeViaOneTime:
		return &render.ValidationError{
			Message: err.Error(),
			Field:   "upgrade_subs_via_one_time",
			Code:    render.CodeInvalid,
		}

	case ErrUnknownPaymentMethod:
		return &render.ValidationError{
			Message: err.Error(),
			Field:   "payment_method",
			Code:    render.CodeInvalid,
		}
	}

	return err
}
