package wxoauth

import (
	"github.com/jmoiron/sqlx"
)

type Env struct {
	db *sqlx.DB
}

func NewEnv(db *sqlx.DB) Env {
	return Env{
		db: db,
	}
}
