package striperepo

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/stripe/stripe-go"
)

// WebHookSaveStripeSub saved a user's membership derived from
// stripe subscription data.
func (env StripeEnv) WebHookOnSubscription(memberID reader.MemberID, ss *stripe.Subscription) error {

	log := logger.WithField("trace", "StripeEnv.OnSubscription")

	tx, err := env.beginOrderTx()
	if err != nil {
		log.Error(err)
		return err
	}

	m, err := tx.RetrieveMember(memberID)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return nil
	}

	if m.PaymentMethod != enum.PayMethodStripe {
		_ = tx.Rollback()
		return sql.ErrNoRows
	}

	if m.StripeSubID.String != ss.ID {
		_ = tx.Rollback()
		return sql.ErrNoRows
	}

	m = ftcStripe.RefreshMembership(m, ss)

	log.Infof("updating a stripe membership: %+v", m)

	// update member
	if err := tx.UpdateMember(m); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
