package giftrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/txrepo"
)

var logger = logrus.WithField("package", "gift_repo")

type GiftEnv struct {
	util.BuildConfig
	db *sqlx.DB
}

func NewGiftEnv(db *sqlx.DB, config util.BuildConfig) GiftEnv {
	return GiftEnv{
		BuildConfig: config,
		db:          db,
	}
}

func (env GiftEnv) beginOrderTx() (txrepo.OrderTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.OrderTx{}, err
	}

	return txrepo.NewOrderTx(tx, env.UseSandboxDB()), nil
}
