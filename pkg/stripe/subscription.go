package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/stripe/stripe-go/v71"
)

// Subscription is a reduced version of stripe.Subscription.
// Used as response when client asks for subscription data.
// Deprecated
type Subscription struct {
	AccountID reader.MemberID `json:"-"`
	product.Edition
	CancelAtPeriodEnd  bool        `json:"cancelAtPeriodEnd"`
	Created            chrono.Time `json:"created"`
	CurrentPeriodEnd   chrono.Time `json:"currentPeriodEnd"`
	CurrentPeriodStart chrono.Time `json:"currentPeriodStart"`
	CustomerID         string      `json:"customerId"`
	EndedAt            chrono.Time `json:"endedAt"`
	ID                 string      `json:"id"`
	LatestInvoiceID    string      `json:"latestInvoiceId"`
	Livemode           bool        `json:"livemode"`
	StartDate          chrono.Time `json:"startDate"`

	// Possible values are incomplete, incomplete_expired, trialing, active, past_due, canceled, or unpaid.
	Status enum.SubsStatus `json:"status"`
}

// NewSubs creates a new Subscription.
// Deprecated.
func NewSubs(s Subs) Subscription {
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

func IsAutoRenewal(s *stripe.Subscription) bool {
	return s.CancelAt == 0 && s.CanceledAt == 0 && !s.CancelAtPeriodEnd
}
