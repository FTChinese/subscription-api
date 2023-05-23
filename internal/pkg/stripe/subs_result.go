package stripe

import (
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// SubsSuccess contains the data after successfully created/modified a subscription:
// - initial creation / upgrade / downgrade / switch-interval
// - cancel or reactivation
// - webhook
type SubsResult struct {
	Modified         bool                       `json:"-"` // A flag to avoid visit db frequently.
	Subs             Subs                       `json:"subs"`
	Member           reader.Membership          `json:"membership"` // New membership.
	Versioned        reader.MembershipVersioned `json:"-"`          // If AnteChange is zero, update Member; otherwise insert Member.
	CarryOverInvoice invoice.Invoice            `json:"-"`          // Only exists when user shifting from one-time purchase to Stripe. It does not exist upon refreshing/upgrading/cancellation.
}
