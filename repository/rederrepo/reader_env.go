package rederrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

var logger = logrus.
	WithField("project", "subscription-api").
	WithField("package", "reader_repo")

type ReaderEnv struct {
	db *sqlx.DB
}

func NewReaderEnv(db *sqlx.DB) ReaderEnv {
	return ReaderEnv{
		db: db,
	}
}

func (env ReaderEnv) FindAccountByFtcID(id string) (reader.Account, error) {
	var u reader.Account
	err := env.db.Get(
		&u,
		query.BuildSelectReader(false, false),
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
		query.BuildSelectReader(true, false),
		cusID)

	if err != nil {
		logger.WithField("trace", "ReaderEnv.FindAccountByStripeID").Error(err)

		return u, err
	}

	return u, nil
}
