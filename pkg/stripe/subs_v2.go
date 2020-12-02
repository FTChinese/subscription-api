package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
)

type SubsV2 struct {
	ID string `json:"id" db:"id"`

	// A date in the future at which the subscription will automatically get canceled
	CancelAtUtc chrono.Time `json:"cancelAtUtc" db:"cancel_at_utc"`
	// If the subscription has been canceled with the at_period_end flag set to true, cancel_at_period_end on the subscription will be true. You can use this attribute to determine whether a subscription that has a status of active is scheduled to be canceled at the end of the current period.
	CancelAtPeriodEnd bool `json:"cancelAtPeriodEnd" db:"cancel_at_period_end"`
	// If the subscription has been canceled, the date of that cancellation. If the subscription was canceled with cancel_at_period_end, canceled_at will reflect the time of the most recent update request, not the end of the subscription period when the subscription is automatically moved to a canceled state.
	CanceledUTC chrono.Time `json:"canceledUtc" db:"canceled_utc"`

	// End of the current period that the subscription has been invoiced for. At the end of this period, a new invoice will be created.
	CurrentPeriodEnd chrono.Time `json:"currentPeriodEnd" db:"current_period_end"`
	// Start of the current period that the subscription has been invoiced for.
	CurrentPeriodStart   chrono.Time `json:"current_period_start" db:"current_period_start"`
	CustomerID           string      `json:"customerId" db:"customer_id"`
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

func NewSubsV2(ftcID string, ss *stripe.Subscription) SubsV2 {
	if ss == nil {
		return SubsV2{}
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

	return SubsV2{
		ID:                   ss.ID,
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
		FtcUserID:            null.StringFrom(ftcID),
	}
}
