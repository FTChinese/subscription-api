package stripe

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"strings"
)

type SharedParams struct {
	CouponID             null.String `json:"coupon"`
	DefaultPaymentMethod null.String `json:"defaultPaymentMethod"`
	IdempotencyKey       string      `json:"idempotency"` // TODO: add when and why this is needed.
}

// SubsInput is the request body to create a new subscription.
type SubsInput struct {
	price.Edition // Use this to find strip's price id.
	SharedParams
}

// Validate checks if customer and idempotency fields are set.
func (i SubsInput) Validate(isUpgrade bool) *render.ValidationError {
	if ve := i.Edition.Validate(); ve != nil {
		return ve
	}

	if isUpgrade && i.Tier != enum.TierPremium {
		return &render.ValidationError{
			Message: "Not the correct edition to upgrade to",
			Field:   "tier",
			Code:    render.CodeInvalid,
		}
	}

	return validator.New("idempotency").Required().Validate(i.IdempotencyKey)
}

// SubsParams contains the data used to create/upgrade a subscription.
type SubsParams struct {
	Account reader.FtcAccount
	Edition price.StripeEdition
	Params  SharedParams
}

type CancelParams struct {
	FtcID  string
	SubID  string
	Cancel bool // True for cancel, false for reactivation.
}

type CheckoutInput struct {
	price.Edition
	SuccessURL string `json:"successUrl"`
	CancelURL  string `json:"cancelUrl"`
}

func (i *CheckoutInput) Validate() *render.ValidationError {
	if ve := i.Edition.Validate(); ve != nil {
		return ve
	}

	i.SuccessURL = strings.TrimSpace(i.SuccessURL)
	i.CancelURL = strings.TrimSpace(i.CancelURL)

	ve := validator.New("successUrl").Required().Validate(i.SuccessURL)
	if ve != nil {
		return ve
	}

	ve = validator.New("cancelUrl").Required().Validate(i.CancelURL)
	if ve != nil {
		return ve
	}

	return nil
}

type CheckoutParams struct {
	Account reader.FtcAccount
	Plan    price.StripeEdition
	Input   CheckoutInput
}
