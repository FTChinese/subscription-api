package txrepo

import (
	"github.com/FTChinese/subscription-api/pkg/giftcard"
	"github.com/jmoiron/sqlx"
)

type GiftCardTx struct {
	SharedTx
}

func NewGiftCardTx(tx *sqlx.Tx) GiftCardTx {
	return GiftCardTx{
		SharedTx: NewSharedTx(tx),
	}
}

func (tx GiftCardTx) ActivateGiftCard(code string) error {
	_, err := tx.Exec(
		giftcard.StmtActivateGiftCard,
		code)

	if err != nil {
		return err
	}

	return nil
}
