package giftrepo

import (
	"github.com/FTChinese/subscription-api/pkg/giftcard"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env Env) RedeemGiftCard(c giftcard.GiftCard, m reader.Membership) error {
	tx, err := env.beginGiftCardTx()
	if err != nil {
		return err
	}

	// Flag the gift card as used.
	err = tx.ActivateGiftCard(c.Code)

	if err != nil {
		_ = tx.Rollback()

	}

	// Insert a new membership.
	err = tx.CreateMember(m)

	if err != nil {
		_ = tx.Rollback()

		// Needs this message to tell client whether
		// there is a duplicate error.
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
