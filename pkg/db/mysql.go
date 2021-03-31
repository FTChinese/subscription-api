package db

import (
	"fmt"
	"github.com/FTChinese/go-rest/connect"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"time"
)

func NewMySQL(c connect.Connect) (*sqlx.DB, error) {
	cfg := &mysql.Config{
		User:   c.User,
		Passwd: c.Pass,
		Net:    "tcp",
		Addr:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		// Always use UTC time.
		// Pay attention to how string values are specified.
		// The string value provided to MySQL must be quoted in single quote for this driver to work,
		// which means the single quote itself must be included in the string value.
		// The resulting string value passed to MySQL should look like: `%27<you string value>%27`
		// See ASCII Encoding Reference https://www.w3schools.com/tags/ref_urlencode.asp
		Params: map[string]string{
			"time_zone": `'+00:00'`,
		},
		Collation:            "utf8mb4_unicode_ci",
		AllowNativePasswords: true,
	}

	db, err := sqlx.Open("mysql", cfg.FormatDSN())

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// When connecting to production server it throws error:
	// packets.go:36: unexpected EOF
	//
	// See https://github.com/go-sql-driver/mysql/issues/674
	db.SetConnMaxLifetime(time.Second)
	return db, nil
}

func MustNewMySQL(c connect.Connect) *sqlx.DB {
	db, err := NewMySQL(c)
	if err != nil {
		panic(err)
	}

	return db
}

type ReadWriteSplit struct {
	Read   *sqlx.DB
	Write  *sqlx.DB
	Delete *sqlx.DB
}

func NewMyDB(prod bool) ReadWriteSplit {
	return ReadWriteSplit{
		Read:   MustNewMySQL(MustMySQLReadConn(prod)),
		Write:  MustNewMySQL(MustMySQLWriteConn(prod)),
		Delete: MustNewMySQL(MustMySQLDeleteConn(prod)),
	}
}
