package db

import "github.com/go-sql-driver/mysql"

// IsAlreadyExists tests if an error means the field already exists
func IsAlreadyExists(err error) bool {
	if e, ok := err.(*mysql.MySQLError); ok && e.Number == 1062 {
		return true
	}

	return false
}
