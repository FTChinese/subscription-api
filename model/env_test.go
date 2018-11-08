package model

import (
	"database/sql"

	cache "github.com/patrickmn/go-cache"
)

func newDevEnv() Env {
	db, err := sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")

	if err != nil {
		panic(err)
	}

	c := cache.New(cache.DefaultExpiration, 0)

	return Env{DB: db, Cache: c}
}

var devEnv = newDevEnv()
