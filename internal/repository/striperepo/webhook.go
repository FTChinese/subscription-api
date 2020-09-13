package striperepo

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	stripePkg "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/stripe/stripe-go"
)

// WebHookSaveStripeSub saves a user's membership derived from
// stripe subscription data.
func (env Env) WebHookOnSubscription(memberID reader.MemberID, ss *stripe.Subscription) error {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginOrderTx()
	if err != nil {
		sugar.Error(err)
		return err
	}

	m, err := tx.RetrieveMember(memberID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil
	}

	if m.PaymentMethod != enum.PayMethodStripe {
		_ = tx.Rollback()
		return sql.ErrNoRows
	}

	if m.StripeSubsID.String != ss.ID {
		_ = tx.Rollback()
		return sql.ErrNoRows
	}

	m = stripePkg.RefreshMembership(m, ss)

	sugar.Infof("updating a stripe membership: %+v", m)

	// update member
	if err := tx.UpdateMember(m); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}
