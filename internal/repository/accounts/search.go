package accounts

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/account"
)

// IDExists searches if an id exists and returns  boolean
// to indicates its existence.
func (env Env) IDExists(id string) (bool, error) {
	var ok bool
	err := env.DBs.Read.Get(&ok, account.StmtIDExists, id)
	if err != nil {
		return false, err
	}

	return ok, nil
}

// EmailExists tests whether a user's email exists.
func (env Env) EmailExists(email string) (bool, error) {
	var ok bool
	err := env.DBs.Read.Get(&ok, account.StmtEmailExists, email)
	if err != nil {
		return false, err
	}

	return ok, nil
}

// NameExists checks if a username exists the user_name column.
func (env Env) NameExists(name string) (bool, error) {
	var ok bool
	err := env.DBs.Read.Get(&ok, account.StmtNameExists, name)
	if err != nil {
		return false, err
	}

	return ok, nil
}

// SearchByEmail tries to find a user's id by email.
// sql.ErrNoRows is not treated as an error here.
// Check the id to see if user is found.
func (env Env) SearchByEmail(email string) (account.SearchResult, error) {
	var result account.SearchResult
	err := env.DBs.Read.Get(&result, account.StmtSearchByEmail, email)
	if err != nil {
		if err != sql.ErrNoRows {
			return account.SearchResult{}, err
		}
	}

	return result, nil
}

// SearchByMobile tries to find a user's id by mobile.
// sql.ErrNoRows is not treated as an error.
// Check the returned id to see if the user exists or not.
func (env Env) SearchByMobile(mobile string) (account.SearchResult, error) {
	var result account.SearchResult
	err := env.DBs.Read.Get(&result, account.StmtSearchByMobile, mobile)
	if err != nil {
		if err != sql.ErrNoRows {
			return account.SearchResult{}, err
		}
	}

	return result, nil
}
