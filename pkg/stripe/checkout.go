package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v71"
)

type Checkout struct {
	Account reader.FtcAccount
	Plan    Plan
	Params  SubsParams
}

// StripeParams create the parameter required by stripe sdk.
func (c Checkout) StripeParams() *stripe.SubscriptionParams {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
		Customer:          stripe.String(c.Account.StripeID.String),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(c.Plan.PriceID),
			},
		},
	}

	// Expand latest_invoice.
	params.AddExpand("latest_invoice.payment_intent")

	// {
	// "status":400,
	// "message":"Idempotent key length is 0 characters long, which is outside accepted lengths. Idempotent Keys must be 1-255 characters long. If you're looking for a decent generator, try using a UUID defined by IETF RFC 4122.",
	// "request_id":"req_O6zILK5QEVpViw",
	// "type":"idempotency_error"
	// }
	if c.Params.IdempotencyKey != "" {
		params.SetIdempotencyKey(c.Params.IdempotencyKey)
	}

	if c.Params.CouponID.Valid {
		params.Coupon = stripe.String(c.Params.DefaultPaymentMethod.String)
	}

	if c.Params.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(c.Params.DefaultPaymentMethod.String)
	}

	return params
}

// NewSubs collects essentials of a stripe subscription.
func (c Checkout) NewSubs(ss *stripe.Subscription) Subs {
	if ss == nil {
		return Subs{}
	}

	var dpm null.String
	if ss.DefaultPaymentMethod != nil {
		dpm = null.StringFrom(ss.DefaultPaymentMethod.ID)
	}

	var invID string
	if ss.LatestInvoice != nil {
		invID = ss.LatestInvoice.ID
	}

	status, _ := enum.ParseSubsStatus(string(ss.Status))

	return Subs{
		ID:                   ss.ID,
		Edition:              c.Plan.Edition,
		CancelAtUtc:          chrono.TimeFrom(dt.FromUnix(ss.CancelAt)),
		CancelAtPeriodEnd:    ss.CancelAtPeriodEnd,
		CanceledUTC:          chrono.TimeFrom(dt.FromUnix(ss.CanceledAt)),
		CurrentPeriodEnd:     chrono.TimeFrom(dt.FromUnix(ss.CurrentPeriodEnd)),
		CurrentPeriodStart:   chrono.TimeFrom(dt.FromUnix(ss.CurrentPeriodStart)),
		CustomerID:           ss.Customer.ID,
		DefaultPaymentMethod: dpm,
		LatestInvoiceID:      invID,
		LiveMode:             ss.Livemode,
		StartDateUTC:         chrono.TimeFrom(dt.FromUnix(ss.StartDate)),
		EndedUTC:             chrono.TimeFrom(dt.FromUnix(ss.EndedAt)),
		CreatedUTC:           chrono.TimeFrom(dt.FromUnix(ss.Created)),
		UpdatedUTC:           chrono.TimeNow(),
		Status:               status,
		FtcUserID:            null.StringFrom(c.Account.FtcID),
	}
}

// Membership builds membership from a stripe subscription.
func (c Checkout) Membership(s Subs) reader.Membership {
	return reader.Membership{
		MemberID:      c.Account.MemberID(),
		Edition:       c.Plan.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    chrono.DateFrom(s.ExpirationTime().Time),
		PaymentMethod: enum.PayMethodStripe,
		FtcPlanID:     null.String{},
		StripeSubsID:  null.StringFrom(s.ID),
		StripePlanID:  null.StringFrom(c.Plan.PriceID),
		AutoRenewal:   s.AutoRenewal(),
		Status:        s.Status,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
	}.Sync()
}

type CheckoutResult struct {
	PaymentResult                       // Deprecated
	Subs          Subs                  `json:"subs"`
	Payment       PaymentResult         `json:"payment"`
	Member        reader.Membership     `json:"membership"`
	Snapshot      reader.MemberSnapshot `json:"-"`
}
