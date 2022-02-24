package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
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
	IsFromStripe bool   `json:"-"` // Flat indicating whether the data comes from Stripe API
	ID           string `json:"id" db:"id"`
	price.Edition
	// A date in the future at which the subscription will automatically get canceled
	WillCancelAtUtc chrono.Time `json:"cancelAtUtc" db:"cancel_at_utc"`
	// Determine whether an active subscription is scheduled to be canceled at the end of the current period. When you cancel a subscription with `at_period_end=true`, this field will be true.
	CancelAtPeriodEnd bool `json:"cancelAtPeriodEnd" db:"cancel_at_period_end"`
	// If the subscription has been canceled, the date of that cancellation.
	// If the subscription was canceled with cancel_at_period_end,
	// canceled_at will reflect the time of the most recent update request,
	// not the end of the subscription period when the subscription is automatically moved to a canceled state.
	CanceledUTC chrono.Time `json:"canceledUtc" db:"canceled_utc"`
	// Start and End of the current period that the subscription has been invoiced for.
	// At the end of this period, a new invoice will be created.
	CurrentPeriodEnd   chrono.Time `json:"currentPeriodEnd" db:"current_period_end"`
	CurrentPeriodStart chrono.Time `json:"currentPeriodStart" db:"current_period_start"`
	CustomerID         string      `json:"customerId" db:"customer_id"`
	// ID of the default payment method for the subscription.
	// It must belong to the customer associated with the subscription.
	// This takes precedence over default_source.
	// If neither are set, invoices will use the customer’s invoice_settings.default_payment_method or default_source.
	// It differs from customer default payment method which might exist even if this one does.
	DefaultPaymentMethodID null.String `json:"defaultPaymentMethod" db:"default_payment_method_id"`
	// If the subscription has ended, the date the subscription ended
	EndedUTC        chrono.Time  `json:"endedUtc" db:"ended_utc"`
	FtcUserID       null.String  `json:"ftcUserId" db:"ftc_user_id"`
	Items           SubsItemList `json:"items" db:"items"`
	LatestInvoiceID string       `json:"latestInvoiceId" db:"latest_invoice_id"`
	LatestInvoice   Invoice      `json:"-"` // To be saved in a separate table.
	LiveMode        bool         `json:"liveMode" db:"live_mode"`
	// Date when the subscription was first created. The date might differ from the created date due to backdating.
	// This does not exist when refreshing current subscription.
	// Occasionally stripe just does not expand it.
	// Do not rely on it.
	PaymentIntentID null.String     `json:"-" db:"payment_intent_id"`
	PaymentIntent   PaymentIntent   `json:"paymentIntent"` // To be saved in a separate table.
	StartDateUTC    chrono.Time     `json:"startDateUtc" db:"start_date_utc"`
	Status          enum.SubsStatus `json:"status" db:"sub_status"`
	// Time at which the object was created. Measured in seconds since the Unix epoch.
	Created int64 `json:"-" db:"created"`

	ItemID string `json:"subsItemId"` // Deprecated. Kept for client deserialization. Will be removed in v7
}

// NewSubs converts stripe's subscription. It returns error if there's
// no subscription item, which should deny membership modification since we have no idea what the user has subscribed to.
// Such kind of error won't happen as long as stripe works.
func NewSubs(ftcID string, ss *stripe.Subscription) Subs {

	var dpm null.String
	if ss.DefaultPaymentMethod != nil {
		dpm = null.StringFrom(ss.DefaultPaymentMethod.ID)
	}

	items := NewSubsItemList(ss.Items)
	var edition price.Edition
	if len(items) > 0 {
		edition = items[0].Price.Edition()
	}

	status := getStatus(ss.Status)

	var pi PaymentIntent
	var inv Invoice
	// LatestInvoice exists even if it is not expanded;
	// however, all pointer fields under invoice will be nil
	if ss.LatestInvoice != nil {
		inv = NewInvoice(ss.LatestInvoice)
		pi = NewPaymentIntent(ss.LatestInvoice.PaymentIntent)
	}

	return Subs{
		IsFromStripe:           true,
		ID:                     ss.ID,
		Edition:                edition,
		WillCancelAtUtc:        chrono.TimeFrom(dt.FromUnix(ss.CancelAt)),
		CancelAtPeriodEnd:      ss.CancelAtPeriodEnd,
		CanceledUTC:            chrono.TimeFrom(dt.FromUnix(ss.CanceledAt)),
		CurrentPeriodEnd:       chrono.TimeFrom(dt.FromUnix(ss.CurrentPeriodEnd)),
		CurrentPeriodStart:     chrono.TimeFrom(dt.FromUnix(ss.CurrentPeriodStart)),
		CustomerID:             ss.Customer.ID,
		DefaultPaymentMethodID: dpm,
		EndedUTC:               chrono.TimeFrom(dt.FromUnix(ss.EndedAt)),
		FtcUserID:              null.NewString(ftcID, ftcID != ""),
		Items:                  items,
		LatestInvoiceID:        inv.ID,
		LatestInvoice:          inv,
		LiveMode:               ss.Livemode,
		PaymentIntentID:        null.NewString(pi.ID, pi.ID == ""),
		PaymentIntent:          pi,
		StartDateUTC:           chrono.TimeFrom(dt.FromUnix(ss.StartDate)),
		Status:                 status,
		Created:                ss.Created,
	}
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

func (s Subs) WithFtcID(id string) Subs {
	s.FtcUserID = null.StringFrom(id)

	return s
}
