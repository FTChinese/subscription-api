package paywall

import (
	"errors"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
)

// StripeSubParams defines the payload when requesting to create/upgrade a stripe subscription.
type StripeSubParams struct {
	Coordinate
	Customer             string      `json:"customer"`
	Coupon               null.String `json:"coupon"`
	DefaultPaymentMethod null.String `json:"defaultPaymentMethod"`
	IdempotencyKey       string      `json:"idempotency"`
	planID               string
}

func (p *StripeSubParams) SetStripePlanID(live bool) error {
	plan, err := GetFtcPlans(live).FindPlan(p.NamedKey())
	if err != nil {
		return nil
	}

	p.planID = plan.StripeID

	return nil
}

func (p StripeSubParams) GetStripePlanID() string {
	return p.planID
}

type StripePayResponse struct {
	RequiresAction            bool        `json:"requiresAction"`
	PaymentIntentClientSecret null.String `json:"paymentIntentClientSecret"`
}

// BuildStripeSubResponse returns the a subscription's payment intent result to client.
func BuildStripeSubResponse(s *stripe.Subscription) (StripePayResponse, error) {
	if s.LatestInvoice == nil {
		return StripePayResponse{}, errors.New("latest_invoice not expanded")
	}

	pi := s.LatestInvoice.PaymentIntent
	if pi == nil {
		return StripePayResponse{}, errors.New("latest_invoice.payment_intent not found")
	}

	return StripePayResponse{
		RequiresAction:            pi.Status == stripe.PaymentIntentStatusRequiresAction,
		PaymentIntentClientSecret: null.StringFrom(pi.ClientSecret),
	}, nil
}
