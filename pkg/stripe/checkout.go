package stripe

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/validator"
	"github.com/guregu/null"
)

type SubsParams struct {
	CouponID             null.String `json:"coupon"`
	DefaultPaymentMethod null.String `json:"defaultPaymentMethod"`
	IdempotencyKey       string      `json:"idempotency"` // TODO: add when and why this is needed.
}

// SubsInput is the request body to create a new subscription.
type SubsInput struct {
	product.Edition // Use this to find strip's price id.
	SubsParams

	PlanID     string `json:"-"`        // Deprecated
	FtcID      string `json:"-"`        // Deprecated.
	CustomerID string `json:"customer"` // Deprecated.
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

// PaymentConfig contains the data used to create/upgrade a subscription.
type PaymentConfig struct {
	Account reader.FtcAccount
	Plan    Plan
	Params  SubsParams
}

type CancelParams struct {
	FtcID  string
	SubID  string
	Cancel bool // True for cancel, false for reactivation.
}
