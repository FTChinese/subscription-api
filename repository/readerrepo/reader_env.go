package readerrepo

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var logger = logrus.
	WithField("project", "subscription-api").
	WithField("package", "reader_repo")

type Env struct {
	config.BuildConfig
	db *sqlx.DB
}

func NewReaderEnv(db *sqlx.DB, config config.BuildConfig) Env {
	return Env{
		BuildConfig: config,
		db:          db,
	}
}

func (env Env) FindAccountByFtcID(id string) (reader.Account, error) {
	var u reader.Account
	err := env.db.Get(
		&u,
		reader.StmtAccountByFtcID,
		id,
	)

	if err != nil {
		logger.WithField("trace", "Env.FindAccountByFtcID").Error(err)

		return u, err
	}

	return u, nil
}

func (env Env) FindAccountByStripeID(cusID string) (reader.Account, error) {
	var u reader.Account
	err := env.db.Get(&u,
		reader.StmtAccountByStripeID,
		cusID)

	if err != nil {
		logger.WithField("trace", "Env.FindAccountByStripeID").Error(err)

		return u, err
	}

	return u, nil
}
