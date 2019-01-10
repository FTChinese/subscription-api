package wepay

import "database/sql"

// Env wraps DB.
type Env struct {
	DB *sql.DB
}
