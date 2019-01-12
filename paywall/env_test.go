package paywall

import (
	"database/sql"

	cache "github.com/patrickmn/go-cache"
)

func newDevEnv() Env {
	db, err := sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")

	if err != nil {
		panic(err)
	}

	return Env{
		DB:    db,
		Cache: cache.New(cache.DefaultExpiration, 0),
	}
}

var devEnv = newDevEnv()
