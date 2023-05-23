package stripe

import (
	"errors"

	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// SubsSyncBuilder collects all the data used to sync
// Stripe subscription into membership.
// You need to take into various factors in this complicated process:
// * Is current membership comes from the same Stripe subscription?
// * Is current membership a one-time purchase? Should Stripe override it?
// * Is current membership comes from Apple?
// We might try to retrieve current membership twice:
// 1. By ftc uuid
// 2. By stripe subscription id.
// However we cannot retrieve both simutaneously since
// this will result dead lock.
// So we first try to retrieve by Stripe subscription.
// If found, it indicates Stripe subscripiton exists;
// otherwise this use might be using one-time purchase
// or Apple.
// Therefore StripeMember and FtcMember are mutually
// exclusive.
// In rare cases both memeberships might be empty.
type SubsSyncBuilder struct {
	UserIDs      ids.UserIDs
	Subs         Subs              // Latest subscription retrieved from Stripe API.
	StripeMember reader.Membership // Current stripe membership.
	FtcMember    reader.Membership // Currernt non-stripe membership. Might be ali/wx/apple
	Archiver     reader.Archiver   // Who performed archiving action.
}

// CurrentMember gets current membership we found in db.
func (b SubsSyncBuilder) CurrentMember() reader.Membership {
	// Always returns the stripe side if exists
	if !b.StripeMember.IsZero() {
		return b.StripeMember
	}

	// Fallback to ftc side. This one might also be empty.
	return b.FtcMember
}

// Build generates WebhookSubsResult.
// Possibilities:
// * No membership corresponds to stripe subscription id. In such case what you can do depends on the ftc side:
//   - Ftc side has no membership. You are safe to create a stripe membership directly;
//   - Ftc side has membership but expired. You can override it.
//   - Ftc side is a one-time purchase, possibilities are:
//   - Stripe subscription expired, you are not allowed to touch ftc membership
//   - Stripe subscription is not expired, override ftc side with carry over addon.
//
// * Stripe side has membership:
//   - If its user id does not match the account retrieve using customer id, it indicates data inconsistency, stop;
//   - Otherwise the stripe membership already exists, simply update it.
//
// Error handling:
//
// * ErrOverrideValidMember - in such case caller should usually save stripe subscripiton to db without touching anything else.
func (b SubsSyncBuilder) Build() (SubsResult, error) {
	// If there's no stripe membership currently.
	// This could happen if:
	// * A new subscription
	// * User already changed from stripe to other methods.
	if b.StripeMember.IsZero() {
		// If ftc side does not have membership.
		// This indicates a new subscription.
		if b.FtcMember.IsZero() {

			newMmb := b.Subs.BuildMembership(
				b.UserIDs,
				addon.AddOn{})

			// Memvership modified.
			return SubsResult{
				Modified: true,
				Subs:     b.Subs,
				Member:   newMmb,
				Versioned: reader.NewMembershipVersioned(newMmb).
					ArchivedBy(b.Archiver),
				CarryOverInvoice: invoice.Invoice{},
			}, nil
		}

		// Ftc side has expired membership.
		// It's safe to override it.
		if b.FtcMember.IsExpired() {

			newMmb := b.Subs.BuildMembership(
				b.UserIDs,
				addon.AddOn{})

			// Membership modified.
			return SubsResult{
				Modified: true,
				Subs:     b.Subs,
				Member:   newMmb,
				Versioned: reader.NewMembershipVersioned(newMmb).
					WithPriorVersion(b.FtcMember).
					ArchivedBy(b.Archiver),
				CarryOverInvoice: invoice.Invoice{},
			}, nil
		}

		// Ftc side is not expired.
		// If stripe subs is valid, permit stripe to only
		// override a one-time payment.
		if b.FtcMember.IsOneTime() {
			// If the subscription refreshed
			// from Stripe API is expired,
			// we shouldn't allow it to override any
			// existing data.
			// Howerver, we should save this subscription without touching
			// anything else.
			// We cannot simply return an error
			// since there's valid partial data.
			if b.Subs.IsExpired() {
				return SubsResult{
					Modified: false,
					Subs:     b.Subs,
					Member:   b.FtcMember,
				}, nil
			}

			// We have both valid Stripe subscription and one-time purchase.
			// With Stripe having a higher weights, override the one-time
			// purchase and turn remaining time of one-time purchase to
			// carry-over add-on.
			inv := b.FtcMember.CarryOverInvoice()
			newMmb := b.Subs.BuildMembership(
				b.UserIDs,
				b.FtcMember.NextRoundAddOn(inv))

			// Membership modified.
			return SubsResult{
				Modified: true,
				Subs:     b.Subs,
				Member:   newMmb,
				Versioned: reader.NewMembershipVersioned(newMmb).
					WithPriorVersion(b.FtcMember).
					ArchivedBy(b.Archiver),
				CarryOverInvoice: inv,
			}, nil
		}

		// The ftc side is still valid and might come from
		// another stripe subscription or apple.
		// Overriding a non-onetime valid membership should be
		// forbidden.
		return SubsResult{
			Modified: false,
			Subs:     b.Subs,
			Member:   b.FtcMember,
		}, nil
	}

	// From now on you should not use FtcMember field
	// any more since it is empty.
	// Since stripe membership exists, it must match user id;
	// otherwise it indicates we are retrieving different users
	// by subscription id and customer id, which shouldn't
	// happen.
	if b.StripeMember.CompoundID != b.UserIDs.CompoundID {
		return SubsResult{}, errors.New("stripe subscription is not targeting this user")
	}

	// Current membership must be created from the subs,
	// simply update it.
	newMmb := b.Subs.BuildMembership(
		b.UserIDs,
		b.StripeMember.AddOn)

	return SubsResult{
		Modified: newMmb.IsModified(b.StripeMember),
		Subs:     b.Subs,
		Member:   newMmb,
		Versioned: reader.NewMembershipVersioned(newMmb).
			WithPriorVersion(b.StripeMember).
			ArchivedBy(b.Archiver),
		CarryOverInvoice: invoice.Invoice{},
	}, nil
}
