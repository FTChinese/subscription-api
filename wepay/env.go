package wepay

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
)

// Env wraps DB.
type Env struct {
	DB *sql.DB
}

var logger = log.WithField("package", "subscription-api.wepay")
