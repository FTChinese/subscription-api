package stripe

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/guregu/null"
)

// SubSharedParams contains common fields of subscription.
type SubSharedParams struct {
	CouponID             null.String `json:"coupon"`
	DefaultPaymentMethod null.String `json:"defaultPaymentMethod"`
	IdempotencyKey       string      `json:"idempotency"` // TODO: add when and why this is needed.
}

type IntroductoryParams struct {
	PriceID    string `json:"priceId"`
	PeriodDays int    `json:"periodDays"`
}

// SubsParams is the request body to create a new subscription
// or update an existing one.
type SubsParams struct {
	PriceID      string             `json:"priceId"`
	Introductory IntroductoryParams `json:"introductory"`
	SubSharedParams
}

// Validate checks if customer and idempotency fields are set.
func (p SubsParams) Validate() *render.ValidationError {
	ve := validator.New("priceId").Required().Validate(p.PriceID)
	if ve != nil {
		return ve
	}

	if p.Introductory.PriceID != "" {
		if p.Introductory.PeriodDays < 1 {
			return &render.ValidationError{
				Message: "Introductory offer period must be provided",
				Field:   "introductory.periodDays",
				Code:    render.CodeInvalid,
			}
		}
	}

	return validator.New("idempotency").Required().Validate(p.IdempotencyKey)
}
