package striperepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/reader"
	stripePkg "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/stripe/stripe-go"
)

// WebHookSaveStripeSub saves a user's membership derived from
// stripe subscription data.
func (env Env) WebHookOnSubscription(memberID reader.MemberID, ss *stripe.Subscription) (reader.Membership, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginOrderTx()
	if err != nil {
		sugar.Error(err)
		return reader.Membership{}, err
	}

	m, err := tx.RetrieveMember(memberID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.Membership{}, err
	}

	// If current membership's payment method is not stripe, discard this notification.
	if !m.IsStripe() {
		_ = tx.Rollback()
		return reader.Membership{}, sql.ErrNoRows
	}

	if m.StripeSubsID.String != ss.ID {
		_ = tx.Rollback()
		return reader.Membership{}, sql.ErrNoRows
	}

	m = stripePkg.RefreshMembership(m, ss)

	sugar.Infof("updating stripe membership from webhook: %+v", m)

	// update member
	if err := tx.UpdateMember(m); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.Membership{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return reader.Membership{}, err
	}

	return m, nil
}
