package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/stripe/stripe-go/v72"
)

// WebHookOnSubscription saves a user's membership derived from
// stripe subscription data.
// TARGET: save the subscription to db whether membership exists or not.
func (env Env) WebHookOnSubscription(account reader.FtcAccount, ss *stripe.Subscription) (reader.MemberSnapshot, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar().
		With("webhook", "stripe-subscription").
		With("id", ss.ID)

	// Build ftc style subscription.
	subs, err := ftcStripe.NewSubs(ss, account.MemberID())
	if err != nil {
		sugar.Error(err)
		return reader.MemberSnapshot{}, err
	}

	// Upsert the subscription. If error occurred, ignore.
	if err := env.UpsertSubs(subs); err != nil {
		sugar.Error(err)
	}

	// Build the new membership.
	newMmb := subs.Membership(ss, account.MemberID())
	sugar.Infof("Updated stripe membership from webhook: %v", newMmb)

	tx, err := env.beginSubsTx()
	if err != nil {
		sugar.Error(err)
		return reader.MemberSnapshot{}, err
	}

	// Retrieve current membership by ftc id.
	// If current membership is empty, we should create it.
	currMmb, err := tx.RetrieveMember(newMmb.MemberID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, err
	}

	// Ensure that current membership is create via stripe.
	if !subs.ShouldUpsert(currMmb) {
		_ = tx.Rollback()
		sugar.Infof("Stripe subscription cannot update/insert its membership")
		return reader.MemberSnapshot{}, nil
	}

	// If nothing changed.
	if !newMmb.IsModified(currMmb) {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, nil
	}

	// Insert to update membership.
	if currMmb.IsZero() {
		if err := tx.CreateMember(newMmb); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return reader.MemberSnapshot{}, err
		}
	} else {
		if err := tx.UpdateMember(newMmb); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return reader.MemberSnapshot{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return reader.MemberSnapshot{}, err
	}

	return currMmb.Snapshot(reader.StripeArchiver(reader.ActionWebhook)), nil
}
