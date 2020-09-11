package wxoauth

import (
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type Env struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Env {
	return Env{
		db: db,
	}
}

var log = logrus.WithField("package", "repository.wxoauth")
