package model

import (
	"database/sql"
)

// Env wraps database connection
type Env struct {
	DB *sql.DB
}
