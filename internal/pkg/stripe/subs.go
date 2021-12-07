package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"time"
)

func getStatus(sts stripe.SubscriptionStatus) enum.SubsStatus {
	status, _ := enum.ParseSubsStatus(string(sts))
	return status
}

// Subs contains the essential data of a stripe subscription.
// It is created from stripe's subscription upon initial creation,
// or refresh, or upgrade.
type Subs struct {
	ID string `json:"id" db:"id"`
	price.Edition
	// A date in the future at which the subscription will automatically get canceled
	WillCancelAtUtc chrono.Time `json:"cancelAtUtc" db:"cancel_at_utc"`
	// If the subscription has been canceled with the at_period_end flag set to true,
	// cancel_at_period_end on the subscription will be true.
	// You can use this attribute to determine whether a subscription that has a status of active is scheduled to be canceled at the end of the current period.
	// When this field is true and it is not the end of current period,
	// status is still active.
	CancelAtPeriodEnd bool `json:"cancelAtPeriodEnd" db:"cancel_at_period_end"`
	// If the subscription has been canceled, the date of that cancellation. If the subscription was canceled with cancel_at_period_end, canceled_at will reflect the time of the most recent update request, not the end of the subscription period when the subscription is automatically moved to a canceled state.
	CanceledUTC chrono.Time `json:"canceledUtc" db:"canceled_utc"`

	// End of the current period that the subscription has been invoiced for. At the end of this period, a new invoice will be created.
	CurrentPeriodEnd chrono.Time `json:"currentPeriodEnd" db:"current_period_end"`
	// Start of the current period that the subscription has been invoiced for.
	CurrentPeriodStart chrono.Time `json:"currentPeriodStart" db:"current_period_start"`

	CustomerID string `json:"customerId" db:"customer_id"`
	// ID of the default payment method for the subscription. It must belong to the customer associated with the subscription. This takes precedence over default_source. If neither are set, invoices will use the customer’s invoice_settings.default_payment_method or default_source.
	DefaultPaymentMethod null.String `json:"defaultPaymentMethod" db:"default_payment_method"`
	// If the subscription has ended, the date the subscription ended
	EndedUTC        chrono.Time `json:"endedUtc" db:"ended_utc"`
	FtcUserID       null.String `json:"ftcUserId" db:"ftc_user_id"`
	LatestInvoiceID string      `json:"latestInvoiceId" db:"latest_invoice_id"`
	LiveMode        bool        `json:"liveMode" db:"live_mode"`
	SubsItem
	// Date when the subscription was first created. The date might differ from the created date due to backdating.
	StartDateUTC chrono.Time     `json:"startDateUtc" db:"start_date_utc"`
	Status       enum.SubsStatus `json:"status" db:"sub_status"`
	// This does not exist when refreshing current subscription.
	// Occasionally stripe just does not expand it.
	// Do not rely on it.
	PaymentIntent PaymentIntent `json:"paymentIntent"`
	// Time at which the object was created. Measured in seconds since the Unix epoch.
	CreatedUTC chrono.Time `json:"createdUtc" db:"created_utc"`
	UpdatedUTC chrono.Time `json:"updatedUtc" db:"updated_utc"`
}

// NewSubs converts stripe's subscription. It returns error if there's
// no subscription item, which should deny membership modification since we have no idea what the user has subscribed to.
// Such kind of error won't happen as long as stripe works.
func NewSubs(ss *stripe.Subscription, ids ids.UserIDs) (Subs, error) {

	var dpm null.String
	if ss.DefaultPaymentMethod != nil {
		dpm = null.StringFrom(ss.DefaultPaymentMethod.ID)
	}

	var invID string
	if ss.LatestInvoice != nil {
		invID = ss.LatestInvoice.ID
	}

	subsItem := NewSubsItem(ss.Items)

	status := getStatus(ss.Status)

	var pi PaymentIntent
	// LatestInvoice might be empty if it is not expanded.
	if ss.LatestInvoice != nil {
		pi = NewPaymentIntent(ss.LatestInvoice.PaymentIntent)
	}

	return Subs{
		ID:                   ss.ID,
		Edition:              subsItem.Price.Edition(),
		WillCancelAtUtc:      chrono.TimeFrom(dt.FromUnix(ss.CancelAt)),
		CancelAtPeriodEnd:    ss.CancelAtPeriodEnd,
		CanceledUTC:          chrono.TimeFrom(dt.FromUnix(ss.CanceledAt)),
		CurrentPeriodEnd:     chrono.TimeFrom(dt.FromUnix(ss.CurrentPeriodEnd)),
		CurrentPeriodStart:   chrono.TimeFrom(dt.FromUnix(ss.CurrentPeriodStart)),
		CustomerID:           ss.Customer.ID,
		DefaultPaymentMethod: dpm,
		SubsItem:             subsItem,
		LatestInvoiceID:      invID,
		LiveMode:             ss.Livemode,
		StartDateUTC:         chrono.TimeFrom(dt.FromUnix(ss.StartDate)),
		EndedUTC:             chrono.TimeFrom(dt.FromUnix(ss.EndedAt)),
		CreatedUTC:           chrono.TimeFrom(dt.FromUnix(ss.Created)),
		UpdatedUTC:           chrono.TimeNow(),
		Status:               status,
		FtcUserID:            ids.FtcID,
		PaymentIntent:        pi,
	}, nil
}

// ExpiresAt determines the exact expiration time.
// If subscription is not in `canceled` status, then
// current_period_end will be the final expiration time.
// If subscription is in canceled state, it might be canceled
// automatically (payment stopped) or manually.
// For automatic cancel, canceled_at should be regarded as the final expiration time.
// For manual cancel, the cancel_at_period_end is true. It will
// expire upon current period end.
func (s Subs) ExpiresAt() time.Time {
	// If status is not in canceled state.
	if s.Status != enum.SubsStatusCanceled {
		return s.CurrentPeriodEnd.Time
	}

	// If canceled. Might be automatic cancel.
	// If it is neither scheduled to cancel at period end, nor
	// in a future time, use the canceled_at field.
	if !s.CancelAtPeriodEnd && s.WillCancelAtUtc.IsZero() {
		return s.CanceledUTC.Time
	}

	// If it is scheduled to cancel at period end, use current_period_end
	if s.CancelAtPeriodEnd {
		return s.CurrentPeriodEnd.Time
	}

	// cancel_at is set, use it.
	// This won't happen for our design.
	return s.WillCancelAtUtc.Time
}

func (s Subs) IsAutoRenewal() bool {
	if s.CancelAtPeriodEnd {
		return false
	}

	return s.Status == enum.SubsStatusActive ||
		s.Status == enum.SubsStatusIncomplete ||
		s.Status == enum.SubsStatusTrialing
}

func (s Subs) IsExpired() bool {
	if s.IsAutoRenewal() {
		return false
	}

	expiresAt := s.ExpiresAt()

	return expiresAt.Before(time.Now().Truncate(24 * time.Hour))
}

// ShouldUpsert checks whether stripe subscription should be allowed to
// create/update user's membership.
func (s Subs) ShouldUpsert(m reader.Membership) bool {
	if m.IsZero() {
		return true
	}

	if m.IsStripe() {
		return true
	}

	if s.ID == m.StripeSubsID.String {
		return true
	}

	return false
}