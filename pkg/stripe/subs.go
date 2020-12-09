package stripe

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"time"
)

// IsSubscribed checks if stripe subscription has valid subscription items.
func IsSubscribed(ss *stripe.Subscription) bool {
	return ss.Items != nil && len(ss.Items.Data) > 0
}

func IsAutoRenewal(ss *stripe.Subscription) bool {
	if ss.CancelAtPeriodEnd {
		return false
	}

	return ss.Status == stripe.SubscriptionStatusActive ||
		ss.Status == stripe.SubscriptionStatusIncomplete ||
		ss.Status == stripe.SubscriptionStatusTrialing
}

func GetExpirationTime(ss *stripe.Subscription) time.Time {
	// If status is not in canceled state.
	if ss.Status != stripe.SubscriptionStatusCanceled {
		return dt.FromUnix(ss.CurrentPeriodEnd)
	}

	// If canceled.
	// If it is neither scheduled to cancel at period end, nor
	// in a future time, use the canceled_at field.
	if !ss.CancelAtPeriodEnd && ss.CancelAt == 0 {
		return dt.FromUnix(ss.CanceledAt)
	}

	// If it is scheduled to cancel at period end, use current_period_end
	if ss.CancelAtPeriodEnd {
		return dt.FromUnix(ss.CurrentPeriodEnd)
	}

	// cancel_at is set, use it.
	return dt.FromUnix(ss.CancelAt)
}

func GetStatus(sts stripe.SubscriptionStatus) enum.SubsStatus {
	status, _ := enum.ParseSubsStatus(string(sts))
	return status
}

// GetSubsItem gets the subscription item id and price id from a stripe subscription.
func GetSubsItem(ss *stripe.Subscription) SubsItem {
	if IsSubscribed(ss) {
		return SubsItem{
			ItemID:  ss.Items.Data[0].ID,
			PriceID: ss.Items.Data[0].Price.ID,
		}
	}

	return SubsItem{}
}

type SubsItem struct {
	ItemID  string `json:"subsItemId" db:"subs_item_id"`
	PriceID string `json:"priceId" db:"price_id"`
}

func (i SubsItem) Plan() (Plan, error) {
	if i.PriceID == "" {
		return Plan{}, errors.New("missing price id in stripe subscription")
	}

	return PlanStore.FindByID(i.PriceID)
}

// Subs contains the essential data of a stripe subscription.
// It it created from stripe's subscription upon initial creation,
// or refresh, or upgrade.
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

	Status    enum.SubsStatus `json:"status" db:"subs_status"`
	FtcUserID null.String     `json:"ftcUserId" db:"ftc_user_id"`
}

// NewSubs converts stripe's subscription. It returns error if there's
// no subscription item, which should deny membership modification since we have no idea what the user has subscribed to.
// Such kind of error won't happen as long as stripe works.
func NewSubs(ss *stripe.Subscription, ids reader.MemberID) (Subs, error) {

	var dpm null.String
	if ss.DefaultPaymentMethod != nil {
		dpm = null.StringFrom(ss.DefaultPaymentMethod.ID)
	}

	var invID string
	if ss.LatestInvoice != nil {
		invID = ss.LatestInvoice.ID
	}

	subsItem := GetSubsItem(ss)

	plan, err := subsItem.Plan()
	if err != nil {
		return Subs{}, err
	}

	status := GetStatus(ss.Status)

	return Subs{
		ID:                   ss.ID,
		Edition:              plan.Edition,
		CancelAtUtc:          chrono.TimeFrom(dt.FromUnix(ss.CancelAt)),
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
	}, nil
}

// Membership creates a reader.Membership for a user.
func (s Subs) Membership(ss *stripe.Subscription, ids reader.MemberID) reader.Membership {
	expTime := GetExpirationTime(ss)

	return reader.Membership{
		MemberID:      ids,
		Edition:       s.Edition,
		LegacyTier:    null.IntFrom(reader.GetTierCode(s.Tier)),
		LegacyExpire:  null.IntFrom(expTime.Unix()),
		ExpireDate:    chrono.DateFrom(expTime),
		PaymentMethod: enum.PayMethodStripe,
		FtcPlanID:     null.String{},
		StripeSubsID:  null.StringFrom(s.ID),
		StripePlanID:  null.StringFrom(s.PriceID),
		AutoRenewal:   IsAutoRenewal(ss),
		Status:        s.Status,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
	}
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

// LegacySubscription builds backward compatible data of a subscription.
// Deprecated.
func (s Subs) LegacySubscription() Subscription {
	return Subscription{
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		Created:            s.CreatedUTC,
		CurrentPeriodEnd:   s.CurrentPeriodEnd,
		CurrentPeriodStart: s.CurrentPeriodStart,
		CustomerID:         s.CustomerID,
		EndedAt:            s.EndedUTC,
		ID:                 s.ID,
		LatestInvoiceID:    s.LatestInvoiceID,
		Livemode:           s.LiveMode,
		StartDate:          s.StartDateUTC,
		Status:             s.Status,
	}
}

// SubsBuilder uses the data of a user's subscription to build the data
// to be saved to db.
type SubsBuilder struct {
	IDs reader.MemberID // IDs might comes from user account, or from current membership for refreshing.
	SS  *stripe.Subscription
	// To build membership, the above three fields are enough.

	CurrentMember reader.Membership    // Used for backup.
	Action        reader.ArchiveAction // Who performed the backup.
}

// Build creates Subs, reader.Membership, reader.MemberSnapshot
// based on ftc user id and stripe subscription.
func (b SubsBuilder) Build() (SubsResult, error) {
	subs, err := NewSubs(b.SS, b.IDs)
	if err != nil {
		return SubsResult{}, err
	}

	m := subs.Membership(b.SS, b.IDs)

	// For refreshing, nothing might be changed.
	isModified := m.IsModified(b.CurrentMember)

	// Only create a snapshot if membership exists and is actually modified.
	var snapshot reader.MemberSnapshot
	if !b.CurrentMember.IsZero() && isModified {
		snapshot = b.CurrentMember.Snapshot(reader.StripeArchiver(b.Action))
	}

	pr, err := NewPaymentResult(b.SS)

	return SubsResult{
		Modified:             isModified,
		MissingPaymentIntent: err != nil,
		PaymentResult:        pr,
		Payment:              pr,
		Subs:                 subs,
		Member:               m,
		Snapshot:             snapshot,
	}, nil
}

// SubsResult contains the data to save to db.
type SubsResult struct {
	Modified             bool                  `json:"-"` // Indicate whether membership actually modified.
	MissingPaymentIntent bool                  `json:"-"` // Whether we failed to expanded latest_invoice.payment_intent. It is not required to create/upgrade a subscription, so we should not return an error.
	PaymentResult                              // Deprecated
	Payment              PaymentResult         `json:"payment"` // Tells user to take further action if any.
	Subs                 Subs                  `json:"subs"`
	Member               reader.Membership     `json:"membership"` // New membership.
	Snapshot             reader.MemberSnapshot `json:"-"`          // If Modified is false, this must exists. If Modified is true, its existence depends -- a newly created membership should not produce a snapshot.
}
