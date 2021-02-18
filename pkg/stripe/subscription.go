package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// Subscription is a reduced version of stripe.Subscription.
// Used as response when client asks for subscription data.
// Deprecated
type Subscription struct {
	AccountID reader.MemberID `json:"-"`
	price.Edition
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
