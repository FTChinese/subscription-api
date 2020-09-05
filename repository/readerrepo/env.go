package readerrepo

import (
	"github.com/FTChinese/subscription-api/pkg/config"
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

func NewEnv(db *sqlx.DB, config config.BuildConfig) Env {
	return Env{
		BuildConfig: config,
		db:          db,
	}
}
