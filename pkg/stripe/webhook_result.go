package stripe

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

const StmtInsertWebhookError = `
INSERT INTO premium.stripe_webhook_error
SET id = :id,
	event_type = :event_type,
	err_message = :message,
	current_stripe_membership = :current_stripe_membership,
	current_dest_membership = :current_dest_membership,
	target_user_id = :target_user_id,
	created_utc = :created_utc
`

type WebhookError struct {
	ID                      string                `db:"id"`
	EventType               string                `db:"event_type"`
	Message                 string                `db:"message"`
	CurrentStripeMembership reader.MembershipJSON `db:"current_stripe_membership"`
	CurrentDestMembership   reader.MembershipJSON `db:"current_dest_membership"`
	TargetUserID            string                `db:"target_user_id"`
	CreatedUTC              chrono.Time           `db:"created_utc"`
}

func (e WebhookError) Error() string {
	return e.Message
}

type WebhookResult struct {
	Member           reader.Membership          // New membership.
	Versioned        reader.MembershipVersioned // If AnteChange is zero, update Member; otherwise insert Member.
	CarryOverInvoice invoice.Invoice
}

type WebhookResultBuilder struct {
	Subs         Subs
	UserIDs      ids.UserIDs
	StripeMember reader.Membership
	FtcMember    reader.Membership
}

// Build generates WebhookResult.
// Possibilities:
// * No membership correspond to stripe subscription id. In such case what you can do depends on the ftc side:
//    * Ftc side has no membership. You are safe to create a stripe membership directly;
//    * Ftc side has membership but expired. You can override it.
//    * Ftc side is a one-time purchase, possibilities are:
//        * Stripe subscription expired, you are not allowed to touch ftc membership
//        * Stripe subscription is not expired, override ftc side with carry over addon.
// * Stripe side has membership:
//    * If its user id does not match the account retrieve using customer id, it indicates data inconsistency, stop;
//    * Otherwise the stripe membership already exists, simply update it.
func (b WebhookResultBuilder) Build() (WebhookResult, error) {
	if b.StripeMember.IsZero() {
		// If ftc side does not have membership.
		if b.FtcMember.IsZero() {
			newMmb := NewMembership(MembershipParams{
				UserIDs: b.UserIDs,
				Subs:    b.Subs,
				AddOn:   addon.AddOn{},
			})
			return WebhookResult{
				Member:           newMmb,
				Versioned:        newMmb.Version(reader.NewStripeArchiver(reader.ActionActionWebhook)),
				CarryOverInvoice: invoice.Invoice{},
			}, nil
		}

		// Ftc side has expired membership.
		// It's safe to override it.
		if b.FtcMember.IsExpired() {
			newMmb := NewMembership(MembershipParams{
				UserIDs: b.UserIDs,
				Subs:    b.Subs,
				AddOn:   b.FtcMember.AddOn,
			})
			return WebhookResult{
				Member: newMmb,
				Versioned: newMmb.Version(
					reader.NewStripeArchiver(reader.ActionActionWebhook)).
					WithPriorVersion(b.FtcMember),
				CarryOverInvoice: invoice.Invoice{},
			}, nil
		}

		// Ftc side is not expired.
		// If stripe subs is valid, permit stripe to only
		// override a one-time payment.
		if b.FtcMember.IsOneTime() {
			if b.Subs.IsExpired() {
				return WebhookResult{}, errors.New("expired stripe subscription cannot override a valid ftc membership")
			}

			// Turn current ftc side to carry over.
			inv := b.FtcMember.CarryOverInvoice()
			newMmb := NewMembership(MembershipParams{
				UserIDs: b.UserIDs,
				Subs:    b.Subs,
				AddOn:   b.FtcMember.AddOn.Plus(addon.New(inv.Tier, inv.TotalDays())),
			})

			return WebhookResult{
				Member: newMmb,
				Versioned: newMmb.
					Version(
						reader.NewStripeArchiver(
							reader.ActionActionWebhook,
						)).
					WithPriorVersion(b.FtcMember),
				CarryOverInvoice: inv,
			}, nil
		}

		// The ftc side is still valid and might come from
		// another stripe subscription or apple.
		// Overriding a non-onetime valid membership should be
		// forbidden. We should record such mismatch.
		return WebhookResult{}, errors.New("cannot permit stripe to override an existing membership")
	}

	// Since stripe membership exists, it must match user id;
	// otherwise it indicates we are retrieving different users
	// by subscription id and customer id, which shouldn't
	// happen.
	if b.StripeMember.CompoundID != b.UserIDs.CompoundID {
		return WebhookResult{}, errors.New("stripe subscription is not targeting this user")
	}

	// Current membership must be created from the subs,
	// simply update it.
	newMmb := NewMembership(MembershipParams{
		UserIDs: b.UserIDs,
		Subs:    b.Subs,
		AddOn:   b.StripeMember.AddOn, // Carry on previous addon without touch.
	})

	return WebhookResult{
		Member: newMmb,
		Versioned: newMmb.Version(
			reader.NewStripeArchiver(
				reader.ActionActionWebhook)).
			WithPriorVersion(b.FtcMember),
		CarryOverInvoice: invoice.Invoice{},
	}, nil
}
