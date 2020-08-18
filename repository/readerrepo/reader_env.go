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

type ReaderEnv struct {
	config.BuildConfig
	db *sqlx.DB
}

func NewReaderEnv(db *sqlx.DB, config config.BuildConfig) ReaderEnv {
	return ReaderEnv{
		BuildConfig: config,
		db:          db,
	}
}

func (env ReaderEnv) FindAccountByFtcID(id string) (reader.Account, error) {
	var u reader.Account
	err := env.db.Get(
		&u,
		reader.StmtAccountByFtcID,
		id,
	)

	if err != nil {
		logger.WithField("trace", "ReaderEnv.FindAccountByFtcID").Error(err)

		return u, err
	}

	return u, nil
}

func (env ReaderEnv) FindAccountByStripeID(cusID string) (reader.Account, error) {
	var u reader.Account
	err := env.db.Get(&u,
		reader.StmtAccountByStripeID,
		cusID)

	if err != nil {
		logger.WithField("trace", "ReaderEnv.FindAccountByStripeID").Error(err)

		return u, err
	}

	return u, nil
}
