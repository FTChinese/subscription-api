package wxoauth

import (
	"github.com/jmoiron/sqlx"
)

type Env struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Env {
	return Env{
		db: db,
	}
}
