package model

import (
	"gitlab.com/ftchinese/subscription-api/paywall"
)

func (env Env) FindGiftCard(code string) (paywall.GiftCard, error) {
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

	var c paywall.GiftCard
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

func (env Env) RedeemGiftCard(c paywall.GiftCard, m paywall.Membership) error {
	tx, err := env.db.Begin()
	if err != nil {
		logger.WithField("trace", "RedeemGiftCard").Error(err)
		return err
	}

	// Flag the gift card as used.
	_, updateErr := tx.Exec(
		env.stmtUseGiftCard(),
		c.Code)

	if updateErr != nil {
		_ = tx.Rollback()

		logger.WithField("trace", "RedeemGiftCard").Error(updateErr)
	}

	// Insert a new membership.
	_, createErr := tx.Exec(
		env.stmtInsertMember(),
		m.CompoundID,
		m.UnionID,
		m.FTCUserID,
		m.UnionID,
		m.Tier,
		m.Cycle,
		m.ExpireDate)

	if createErr != nil {
		_ = tx.Rollback()

		logger.WithField("trace", "RedeemGiftCard").Error(err)
		// Needs this message to tell client whether
		// there is a duplicate error.
		return createErr
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("trace", "Redeem").Error(err)
		return err
	}

	return nil
}
