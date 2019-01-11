package util

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

// Shared errors.
var (
	// ErrRenewalForbidden indicates a user is not allowed to renew membership
	ErrRenewalForbidden = errors.New("membership renewal forbidden")
	ErrAlreadyExists    = errors.New("already exists")
	// ErrIncompatible is returned by Scan interface
	ErrIncompatible = errors.New("incompatible type to scan")
)

// IsAlreadyExists tests if SQL error is an duplicate error.
func IsAlreadyExists(err error) bool {
	if e, ok := err.(*mysql.MySQLError); ok && e.Number == 1062 {
		return true
	}

	if err == ErrAlreadyExists {
		return true
	}

	return false
}
