package giftrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/jmoiron/sqlx"
)

type Env struct {
	config.BuildConfig
	db *sqlx.DB
}

func New(db *sqlx.DB, config config.BuildConfig) Env {
	return Env{
		BuildConfig: config,
		db:          db,
	}
}

func (env Env) beginGiftCardTx() (txrepo.GiftCardTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.GiftCardTx{}, err
	}

	return txrepo.NewGiftCardTx(tx), nil
}
