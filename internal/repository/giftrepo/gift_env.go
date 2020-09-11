package giftrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("package", "gift_repo")

type GiftEnv struct {
	config.BuildConfig
	db *sqlx.DB
}

func NewGiftEnv(db *sqlx.DB, config config.BuildConfig) GiftEnv {
	return GiftEnv{
		BuildConfig: config,
		db:          db,
	}
}

func (env GiftEnv) beginOrderTx() (txrepo.MemberTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.MemberTx{}, err
	}

	return txrepo.NewMemberTx(tx), nil
}
