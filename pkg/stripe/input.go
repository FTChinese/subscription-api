package stripe

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"strings"
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

	return validator.New("idempotency").Required().Validate(i.IdempotencyKey)
}

// SubsParams contains the data used to create/upgrade a subscription.
type SubsParams struct {
	Account account.BaseAccount
	Edition price.StripeEdition
	SharedParams
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
	Account account.BaseAccount
	Plan    price.StripeEdition
	Input   CheckoutInput
}
