package reader

import (
	"errors"
	"github.com/FTChinese/go-rest/render"
	"net/http"
)

var (
	ErrNonStripeValidSub  = errors.New("already subscribed via non-stripe method")
	ErrStripeDuplicateSub = errors.New("already an active stripe subscription")
	ErrStripeNoDowngrade  = errors.New("downgrading subscription is not supported currently")
	ErrStripeNotCreate    = errors.New("only creating new subscription is supported by this endpoint")
	ErrUnknownSubState    = errors.New("your subscription status cannot be determined")
)

// ParseStripeSubError turns error to ValidationError, or pass it down as is.
func ParseStripeSubError(err error) (*render.ResponseError, bool) {
	switch err {
	case ErrNonStripeValidSub:
		return &render.ResponseError{
			StatusCode: http.StatusUnprocessableEntity,
			Message:    "You are already subscribed via non-stripe method",
			Invalid: &render.ValidationError{
				Field: "payMethod",
				Code:  render.CodeInvalid,
			},
		}, true

	case ErrStripeDuplicateSub:
		return &render.ResponseError{
			StatusCode: http.StatusUnprocessableEntity,
			Message:    "You are already subscribed via Stripe",
			Invalid: &render.ValidationError{
				Field: "membership",
				Code:  render.CodeAlreadyExists,
			},
		}, true

	case ErrStripeNoDowngrade:
		return &render.ResponseError{
			StatusCode: http.StatusUnprocessableEntity,
			Message:    "Downgrading is not supported currently",
			Invalid: &render.ValidationError{
				Field: "downgrade",
				Code:  render.CodeInvalid,
			},
		}, true

	case ErrStripeNotCreate:
		return &render.ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "This endpoint only support creating new subscription",
			Invalid:    nil,
		}, true

	default:
		return nil, false
	}
}
