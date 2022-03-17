package stripe

import (
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/stripe/stripe-go/v72"
)

// SubsResultParams uses the data of a user's subscription to build the data to be saved to db.
type SubsResultParams struct {
	UserIDs ids.UserIDs // UserIDs might comes from user account, or from current membership for refreshing.
	Kind    reader.SubsKind
	// To build membership, the above three fields are enough.

	CurrentMember reader.Membership    // Used for backup.
	Action        reader.ArchiveAction // Who performed the backup.
}

// SubsResult contains the data to save to db.
type SubsResult struct {
	Modified         bool                       `json:"-"` // Indicate whether membership actually modified.
	Subs             Subs                       `json:"subs"`
	Member           reader.Membership          `json:"membership"` // New membership.
	Versioned        reader.MembershipVersioned `json:"-"`
	CarryOverInvoice invoice.Invoice            `json:"-"` // Only exists when user shifting from one-time purchase to Stripe. It does not exists upon refreshing/upgrading/cancellation.
}

func NewSubsResult(ss *stripe.Subscription, params SubsResultParams) SubsResult {
	return newSubsResult(
		NewSubs(params.UserIDs.FtcID.String, ss),
		params,
	)
}

// newSubsResult exists for testing convenience.
func newSubsResult(subs Subs, params SubsResultParams) SubsResult {

	var inv invoice.Invoice
	if params.Kind == reader.SubsKindOneTimeToSub {
		inv = params.
			CurrentMember.
			CarryOverInvoice().
			WithStripeSubsID(subs.ID)
	}

	newMmb := NewMembership(MembershipParams{
		UserIDs: params.UserIDs,
		Subs:    subs,
		AddOn: params.CurrentMember.
			AddOn.
			Plus(addon.New(inv.Tier, inv.TotalDays())),
	})

	// For refreshing, nothing might be changed; or user might switched
	// to other purchase channel.
	isModified := newMmb.IsModified(params.CurrentMember) && subs.ShouldUpsert(params.CurrentMember)

	// Only create a snapshot if membership is actually modified.
	var versioned reader.MembershipVersioned
	if isModified {
		versioned = newMmb.Version(reader.NewStripeArchiver(params.Action))
	}
	// Keep a previous version of membership only when exists.
	if !params.CurrentMember.IsZero() {
		versioned = versioned.WithPriorVersion(params.CurrentMember)
	}

	return SubsResult{
		Modified:         isModified,
		Subs:             subs,
		Member:           newMmb,
		Versioned:        versioned,
		CarryOverInvoice: inv,
	}
}
