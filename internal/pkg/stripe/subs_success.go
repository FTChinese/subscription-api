package stripe

import (
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// SubsSuccessBuilder uses the data of a user's subscription to build the data to be saved to db.
type SubsSuccessBuilder struct {
	UserIDs       ids.UserIDs // UserIDs might come from user account, or from current membership for refreshing.
	Kind          reader.SubsIntentKind
	CurrentMember reader.Membership // Used for backup.
	Subs          Subs
	Archiver      reader.Archiver
}

func NewSubsBuilder(cart reader.ShoppingCart, subs Subs, archiver reader.Archiver) SubsSuccessBuilder {
	return SubsSuccessBuilder{
		UserIDs:       cart.Account.CompoundIDs(),
		Kind:          cart.Intent.Kind,
		CurrentMember: cart.CurrentMember,
		Subs:          subs,
		Archiver:      archiver,
	}
}

func (b SubsSuccessBuilder) Build() SubsResult {
	return b.build()
}

func (b SubsSuccessBuilder) build() SubsResult {
	var inv invoice.Invoice
	if b.Kind.IsSwitchToAutoRenew() {
		inv = b.
			CurrentMember.
			CarryOverInvoice().
			WithStripeSubsID(b.Subs.ID)
	}

	newMmb := b.Subs.BuildMembership(
		b.UserIDs,
		b.CurrentMember.NextRoundAddOn(inv))

	// For refreshing, nothing might be changed; or user might
	// have already switched to other purchase channel.
	isModified := newMmb.IsModified(b.CurrentMember) && b.Subs.ShouldUpsert(b.CurrentMember)

	// Only create a snapshot if membership is actually modified.
	var versioned reader.MembershipVersioned
	if isModified {
		versioned = reader.NewMembershipVersioned(newMmb).
			WithPriorVersion(b.CurrentMember).
			ArchivedBy(b.Archiver)
	}

	return SubsResult{
		Modified:         isModified,
		Subs:             b.Subs,
		Member:           newMmb,
		Versioned:        versioned,
		CarryOverInvoice: inv,
	}
}
