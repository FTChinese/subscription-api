package repository

import (
	"gitlab.com/ftchinese/subscription-api/models/redeem"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
)

func (env Env) FindGiftCard(code string) (redeem.GiftCard, error) {
	query := `
	SELECT auth_code AS redeemCode,
		tier AS tier,
		cycle_unit AS cycleUnit,
		cycle_value AS cycleValue
	FROM premium.scratch_card
	WHERE auth_code = ?
		AND expire_time > UNIX_TIMESTAMP()
		AND active_time = 0
	LIMIT 1`

	var c redeem.GiftCard
	err := env.db.QueryRow(query, code).Scan(
		&c.Code,
		&c.Tier,
		&c.CycleUnit,
		&c.CycleValue)

	if err != nil {
		logger.WithField("trace", "FindGiftCard").Error(err)
		return c, err
	}

	return c, nil
}

func (env Env) RedeemGiftCard(c redeem.GiftCard, m subscription.Membership) error {
	tx, err := env.BeginOrderTx()
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
