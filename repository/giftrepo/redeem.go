package giftrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/redeem"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
)

func (env GiftEnv) RedeemGiftCard(c redeem.GiftCard, m subscription.Membership) error {
	tx, err := env.beginOrderTx()
	if err != nil {
		logger.WithField("trace", "RedeemGiftCard").Error(err)
		return err
	}

	// Flag the gift card as used.
	err = tx.ActivateGiftCard(c.Code)

	if err != nil {
		_ = tx.Rollback()

		logger.WithField("trace", "RedeemGiftCard").Error(err)
	}

	// Insert a new membership.
	err = tx.CreateMember(m)

	if err != nil {
		_ = tx.Rollback()

		logger.WithField("trace", "RedeemGiftCard").Error(err)
		// Needs this message to tell client whether
		// there is a duplicate error.
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("trace", "Redeem").Error(err)
		return err
	}

	return nil
}
