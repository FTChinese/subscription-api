package db

import (
	"fmt"
	"log"
	"time"

	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func buildDSN(c config.Connect) string {
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

	return cfg.FormatDSN()
}

func NewMySQL(c config.Connect) (*sqlx.DB, error) {

	db, err := sqlx.Open("mysql", buildDSN(c))

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
	log.Printf("Connected to MySQL %s", c.Host)
	return db, nil
}

func MustNewMySQL(c config.Connect) *sqlx.DB {
	db, err := NewMySQL(c)
	if err != nil {
		panic(err)
	}

	return db
}

func MustNewMyDBs() ReadWriteMyDBs {
	return ReadWriteMyDBs{
		Read:   MustNewMySQL(config.MustMySQLReadConn()),
		Write:  MustNewMySQL(config.MustMySQLWriteConn()),
		Delete: MustNewMySQL(config.MustMySQLDeleteConn()),
	}
}

type ReadWriteMyDBs struct {
	Read   *sqlx.DB
	Write  *sqlx.DB
	Delete *sqlx.DB
}
