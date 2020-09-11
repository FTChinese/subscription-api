package readerrepo

import (
	"github.com/jmoiron/sqlx"
)

type Env struct {
	//config.BuildConfig
	db *sqlx.DB
}

func NewEnv(db *sqlx.DB) Env {
	return Env{
		db: db,
	}
}
