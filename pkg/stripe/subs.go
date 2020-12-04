package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
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
	product.Edition
	PlanID string `json:"-"` // Deprecated

	FtcID      string `json:"-"`        // Deprecated. Acquired from header.
	CustomerID string `json:"customer"` // Deprecated. Use FtcID to retrieve user account.

	SubsParams
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

// Subs contains the essential data of a stripe subscription.
type Subs struct {
	ID string `json:"id" db:"id"`
	product.Edition
	// A date in the future at which the subscription will automatically get canceled
	CancelAtUtc chrono.Time `json:"cancelAtUtc" db:"cancel_at_utc"`
	// If the subscription has been canceled with the at_period_end flag set to true, cancel_at_period_end on the subscription will be true. You can use this attribute to determine whether a subscription that has a status of active is scheduled to be canceled at the end of the current period.
	// When this field is true and it is not the end of current period,
	// status is still active.
	CancelAtPeriodEnd bool `json:"cancelAtPeriodEnd" db:"cancel_at_period_end"`
	// If the subscription has been canceled, the date of that cancellation. If the subscription was canceled with cancel_at_period_end, canceled_at will reflect the time of the most recent update request, not the end of the subscription period when the subscription is automatically moved to a canceled state.
	CanceledUTC chrono.Time `json:"canceledUtc" db:"canceled_utc"`

	// End of the current period that the subscription has been invoiced for. At the end of this period, a new invoice will be created.
	CurrentPeriodEnd chrono.Time `json:"currentPeriodEnd" db:"current_period_end"`
	// Start of the current period that the subscription has been invoiced for.
	CurrentPeriodStart chrono.Time `json:"current_period_start" db:"current_period_start"`
	CustomerID         string      `json:"customerId" db:"customer_id"`
	// ID of the default payment method for the subscription. It must belong to the customer associated with the subscription. This takes precedence over default_source. If neither are set, invoices will use the customer’s invoice_settings.default_payment_method or default_source.
	DefaultPaymentMethod null.String `json:"defaultPaymentMethod" db:"default_payment_method"`
	LatestInvoiceID      string      `json:"latestInvoiceId" db:"latest_invoice_id"`
	LiveMode             bool        `json:"liveMode" db:"live_mode"`

	// Date when the subscription was first created. The date might differ from the created date due to backdating.
	StartDateUTC chrono.Time `json:"startDateUtc" db:"start_date_utc"`
	// If the subscription has ended, the date the subscription ended
	EndedUTC chrono.Time `json:"endedUtc" db:"ended_utc"`

	// Time at which the object was created. Measured in seconds since the Unix epoch.
	CreatedUTC chrono.Time `json:"createdUtc" db:"created_utc"`
	UpdatedUTC chrono.Time `json:"updatedUtc" db:"updated_utc"`

	Status    enum.SubsStatus `json:"status" db:"subs_status"`
	FtcUserID null.String     `json:"ftcUserId" db:"ftc_user_id"`
}

func (s Subs) UpdateMember(m reader.Membership) reader.Membership {
	m.ExpireDate = chrono.DateFrom(s.ExpirationTime().Time)
	m.Status = s.Status
	m.AutoRenewal = s.AutoRenewal()

	return m
}

// AutoRenewal determines whether the subscription is auto renewal.
// All subscriptions are auto renewal unless it is in a invalid state.
func (s Subs) AutoRenewal() bool {
	// If you schedule to cancel this subscription at period end,
	// the status is still active.
	if s.CancelAtPeriodEnd {
		return false
	}

	return s.Status.IsValid()
}

func (s Subs) ExpirationTime() chrono.Time {
	// If status is not in canceled state.
	if s.Status != enum.SubsStatusCanceled {
		return s.CurrentPeriodEnd
	}

	// If canceled.
	// If it is neither scheduled to cancel at period end, nor
	// in a future time, use the canceled_at field.
	if !s.CancelAtPeriodEnd && s.CancelAtUtc.IsZero() {
		return s.CancelAtUtc
	}

	// If it is scheduled to cancel at period end, use current_period_end
	if s.CancelAtPeriodEnd {
		return s.CurrentPeriodEnd
	}

	// cancel_at is set, use it.
	return s.CancelAtUtc
}
