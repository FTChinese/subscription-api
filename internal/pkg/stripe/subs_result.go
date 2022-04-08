package stripe

import (
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/stripe/stripe-go/v72"
)

type SubsError struct {
	Membership reader.Membership
	Raw        error
}

func (s SubsError) Error() string {
	return s.Raw.Error()
}

// SubsResultParams uses the data of a user's subscription to build the data to be saved to db.
type SubsResultParams struct {
	UserIDs       ids.UserIDs // UserIDs might comes from user account, or from current membership for refreshing.
	Kind          reader.SubsKind
	CurrentMember reader.Membership    // Used for backup.
	Action        reader.ArchiveAction // Who performed the backup.
}

// SubsSuccess contains the data to save to db.
type SubsSuccess struct {
	Modified         bool                       `json:"-"` // Indicate whether membership actually modified.
	Subs             Subs                       `json:"subs"`
	Member           reader.Membership          `json:"membership"` // New membership.
	Versioned        reader.MembershipVersioned `json:"-"`
	CarryOverInvoice invoice.Invoice            `json:"-"` // Only exists when user shifting from one-time purchase to Stripe. It does not exists upon refreshing/upgrading/cancellation.
}

func NewSubsResult(ss *stripe.Subscription, params SubsResultParams) SubsSuccess {
	return newSubsResult(
		NewSubs(params.UserIDs.FtcID.String, ss),
		params,
	)
}

func NewSubsCreated(cart pw.ShoppingCart, ss *stripe.Subscription) SubsSuccess {
	subs := NewSubs(cart.Account.FtcID, ss)

	return newSubsResult(
		subs,
		SubsResultParams{
			UserIDs:       cart.Account.CompoundIDs(),
			Kind:          cart.Intent.Kind,
			CurrentMember: cart.CurrentMember,
			Action:        reader.ArchiveActionCreate,
		})
}

// newSubsResult exists for testing convenience.
func newSubsResult(subs Subs, params SubsResultParams) SubsSuccess {

	var inv invoice.Invoice
	if params.Kind == reader.SubsKindOneTimeToAutoRenew {
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

	return SubsSuccess{
		Modified:         isModified,
		Subs:             subs,
		Member:           newMmb,
		Versioned:        versioned,
		CarryOverInvoice: inv,
	}
}
