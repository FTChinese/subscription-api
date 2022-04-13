package stripe

import (
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/stripe/stripe-go/v72"
)

// SubsResultBuilder uses the data of a user's subscription to build the data to be saved to db.
type SubsResultBuilder struct {
	UserIDs       ids.UserIDs // UserIDs might come from user account, or from current membership for refreshing.
	Kind          reader.SubsIntentKind
	CurrentMember reader.Membership // Used for backup.
	Archiver      reader.Archiver
}

func NewSubsResultBuilder(cart reader.ShoppingCart, archiver reader.Archiver) SubsResultBuilder {
	return SubsResultBuilder{
		UserIDs:       cart.Account.CompoundIDs(),
		Kind:          cart.Intent.Kind,
		CurrentMember: cart.CurrentMember,
		Archiver:      archiver,
	}
}

func (b SubsResultBuilder) Build(ss *stripe.Subscription) SubsSuccess {
	subs := NewSubs(b.UserIDs.FtcID.String, ss)

	return newSubsResult(
		subs,
		b)
}

// SubsSuccess contains the data after successfully created/modified a subscription:
// - initial creation / upgrade / downgrade / switch-interval
// - cancel or reactivation
// - webhook
type SubsSuccess struct {
	Modified         bool                       `json:"-"` // Indicate whether membership actually modified.
	Subs             Subs                       `json:"subs"`
	Member           reader.Membership          `json:"membership"` // New membership.
	Versioned        reader.MembershipVersioned `json:"-"`
	CarryOverInvoice invoice.Invoice            `json:"-"` // Only exists when user shifting from one-time purchase to Stripe. It does not exist upon refreshing/upgrading/cancellation.
}

// newSubsResult exists for testing convenience.
func newSubsResult(subs Subs, params SubsResultBuilder) SubsSuccess {

	var inv invoice.Invoice
	if params.Kind.IsSwitchToAutoRenew() {
		inv = params.
			CurrentMember.
			CarryOverInvoice().
			WithStripeSubsID(subs.ID)
	}

	newMmb := subs.BuildMembership(
		params.UserIDs,
		params.CurrentMember.NextRoundAddOn(inv))

	// For refreshing, nothing might be changed; or user might
	// have already switched to other purchase channel.
	isModified := newMmb.IsModified(params.CurrentMember) && subs.ShouldUpsert(params.CurrentMember)

	// Only create a snapshot if membership is actually modified.
	var versioned reader.MembershipVersioned
	if isModified {
		versioned = reader.NewMembershipVersioned(newMmb).
			WithPriorVersion(params.CurrentMember).
			ArchivedBy(params.Archiver)
	}

	return SubsSuccess{
		Modified:         isModified,
		Subs:             subs,
		Member:           newMmb,
		Versioned:        versioned,
		CarryOverInvoice: inv,
	}
}
