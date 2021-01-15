package stripe

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/stripe/stripe-go/v72"
)

// SubsResultParams uses the data of a user's subscription to build the data to be saved to db.
type SubsResultParams struct {
	UserIDs reader.MemberID // UserIDs might comes from user account, or from current membership for refreshing.
	SS      *stripe.Subscription
	// To build membership, the above three fields are enough.

	CurrentMember reader.Membership    // Used for backup.
	Action        reader.ArchiveAction // Who performed the backup.
}

// SubsResult contains the data to save to db.
type SubsResult struct {
	Modified             bool                  `json:"-"`       // Indicate whether membership actually modified.
	MissingPaymentIntent bool                  `json:"-"`       // Whether we failed to expanded latest_invoice.payment_intent. It is not required to create/upgrade a subscription, so we should not return an error.
	Payment              PaymentResult         `json:"payment"` // Tells user to take further action if any.
	Subs                 Subs                  `json:"subs"`
	Member               reader.Membership     `json:"membership"` // New membership.
	Snapshot             reader.MemberSnapshot `json:"-"`          // If Modified is false, this must exists. If Modified is true, its existence depends -- a newly created membership should not produce a snapshot.
}

func NewSubsResult(params SubsResultParams) (SubsResult, error) {
	subs, err := NewSubs(params.SS, params.UserIDs)
	if err != nil {
		return SubsResult{}, err
	}

	m := NewMembership(MembershipParams{
		UserIDs:   params.UserIDs,
		Subs:      subs,
		AddOnDays: params.CurrentMember.ReservedDays,
	})

	// For refreshing, nothing might be changed.
	isModified := m.IsModified(params.CurrentMember)

	// Only create a snapshot if membership exists and is actually modified.
	var snapshot reader.MemberSnapshot
	if !params.CurrentMember.IsZero() && isModified {
		snapshot = params.CurrentMember.Snapshot(reader.StripeArchiver(params.Action))
	}

	pr, err := NewPaymentResult(params.SS)

	return SubsResult{
		Modified:             isModified,
		MissingPaymentIntent: err != nil,
		Payment:              pr,
		Subs:                 subs,
		Member:               m,
		Snapshot:             snapshot,
	}, nil
}
