package readerrepo

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/jmoiron/sqlx"
)

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
