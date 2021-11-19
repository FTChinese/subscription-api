package giftrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/jmoiron/sqlx"
)

type Env struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Env {
	return Env{
		db: db,
	}
}

func (env Env) beginGiftCardTx() (txrepo.GiftCardTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.GiftCardTx{}, err
	}

	return txrepo.NewGiftCardTx(tx), nil
}
