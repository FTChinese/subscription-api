package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"time"
)

// isSubscribed checks if stripe subscription has valid subscription items.
func isSubscribed(ss *stripe.Subscription) bool {
	return ss.Items != nil && len(ss.Items.Data) > 0
}

func getStatus(sts stripe.SubscriptionStatus) enum.SubsStatus {
	status, _ := enum.ParseSubsStatus(string(sts))
	return status
}

type SubsItem struct {
	ItemID  string `json:"subsItemId" db:"subs_item_id"`
	PriceID string `json:"priceId" db:"price_id"`
}

// NewSubsItem gets the subscription item id and price id from a stripe subscription.
func NewSubsItem(ss *stripe.Subscription) SubsItem {
	if isSubscribed(ss) {
		return SubsItem{
			ItemID:  ss.Items.Data[0].ID,
			PriceID: ss.Items.Data[0].Price.ID,
		}
	}

	return SubsItem{}
}

// Subs contains the essential data of a stripe subscription.
// It it created from stripe's subscription upon initial creation,
// or refresh, or upgrade.
type Subs struct {
	ID string `json:"id" db:"id"`
	price.Edition
	// A date in the future at which the subscription will automatically get canceled
	WillCancelAtUtc chrono.Time `json:"cancelAtUtc" db:"cancel_at_utc"`
	// If the subscription has been canceled with the at_period_end flag set to true, cancel_at_period_end on the subscription will be true. You can use this attribute to determine whether a subscription that has a status of active is scheduled to be canceled at the end of the current period.
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
	SubsItem
	LatestInvoiceID string `json:"latestInvoiceId" db:"latest_invoice_id"`
	LiveMode        bool   `json:"liveMode" db:"live_mode"`

	// Date when the subscription was first created. The date might differ from the created date due to backdating.
	StartDateUTC chrono.Time `json:"startDateUtc" db:"start_date_utc"`
	// If the subscription has ended, the date the subscription ended
	EndedUTC chrono.Time `json:"endedUtc" db:"ended_utc"`

	// Time at which the object was created. Measured in seconds since the Unix epoch.
	CreatedUTC chrono.Time `json:"createdUtc" db:"created_utc"`
	UpdatedUTC chrono.Time `json:"updatedUtc" db:"updated_utc"`

	Status        enum.SubsStatus `json:"status" db:"sub_status"`
	FtcUserID     null.String     `json:"ftcUserId" db:"ftc_user_id"`
	PaymentIntent PaymentIntent   `json:"paymentIntent"` // This does not exists when refreshing current subscription.
}

// NewSubs converts stripe's subscription. It returns error if there's
// no subscription item, which should deny membership modification since we have no idea what the user has subscribed to.
// Such kind of error won't happen as long as stripe works.
func NewSubs(ss *stripe.Subscription, ids pkg.MemberID) (Subs, error) {

	var dpm null.String
	if ss.DefaultPaymentMethod != nil {
		dpm = null.StringFrom(ss.DefaultPaymentMethod.ID)
	}

	var invID string
	if ss.LatestInvoice != nil {
		invID = ss.LatestInvoice.ID
	}

	subsItem := NewSubsItem(ss)

	plan, err := price.StripeEditions.FindByID(subsItem.PriceID)
	if err != nil {
		return Subs{}, err
	}

	status := getStatus(ss.Status)

	var pi PaymentIntent
	// LatestInvoice might be empty if it is not expanded.
	if ss.LatestInvoice != nil {
		pi = NewPaymentIntent(ss.LatestInvoice.PaymentIntent)
	}

	return Subs{
		ID:                   ss.ID,
		Edition:              plan.Edition,
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

func (s Subs) ExpiresAt() time.Time {
	// If status is not in canceled state.
	if s.Status != enum.SubsStatusCanceled {
		return s.CurrentPeriodEnd.Time
	}

	// If canceled.
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
