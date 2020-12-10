package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	stripeSdk "github.com/stripe/stripe-go/v72"
)

// OnSubscription save stripe subscription and optionally update membership linked to it.
func (env Env) OnSubscription(account reader.FtcAccount, ss *stripeSdk.Subscription) (stripe.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar().
		With("webhook", "stripe-subscription").
		With("id", ss.ID)

	// Build ftc style subscription.
	subs, err := stripe.NewSubs(ss, account.MemberID())
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Build the new membership.
	newMmb := subs.Membership(ss, account.MemberID())
	sugar.Infof("Updated stripe membership from webhook: %v", newMmb)

	tx, err := env.beginSubsTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Retrieve current membership by ftc id.
	// If current membership is empty, we should create it.
	currMmb, err := tx.RetrieveMember(newMmb.MemberID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	// Ensure that current membership is create via stripe.
	if !subs.ShouldUpsert(currMmb) {
		_ = tx.Rollback()
		sugar.Infof("Stripe subscription cannot update/insert its membership")
		return stripe.SubsResult{
			Modified:             false,
			MissingPaymentIntent: false,
			PaymentResult:        stripe.PaymentResult{},
			Payment:              stripe.PaymentResult{},
			Subs:                 subs,
			Member:               currMmb,
			Snapshot:             reader.MemberSnapshot{},
		}, nil
	}

	// If nothing changed.
	if !newMmb.IsModified(currMmb) {
		_ = tx.Rollback()
		return stripe.SubsResult{
			Modified:             false,
			MissingPaymentIntent: false,
			PaymentResult:        stripe.PaymentResult{},
			Payment:              stripe.PaymentResult{},
			Subs:                 subs,
			Member:               newMmb,
			Snapshot:             reader.MemberSnapshot{},
		}, nil
	}

	// Insert to update membership.
	if currMmb.IsZero() {
		if err := tx.CreateMember(newMmb); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsResult{}, err
		}
	} else {
		if err := tx.UpdateMember(newMmb); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	return stripe.SubsResult{
		Modified:             true,
		MissingPaymentIntent: false,
		PaymentResult:        stripe.PaymentResult{},
		Payment:              stripe.PaymentResult{},
		Subs:                 subs,
		Member:               newMmb,
		Snapshot:             currMmb.Snapshot(reader.StripeArchiver(reader.ActionWebhook)),
	}, nil
}
