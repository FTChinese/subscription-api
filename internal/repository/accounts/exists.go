package accounts

import "github.com/FTChinese/subscription-api/pkg/account"

func (env Env) IDExists(id string) (bool, error) {
	var ok bool
	err := env.dbs.Read.Get(&ok, account.StmtIDExists, id)
	if err != nil {
		return false, err
	}

	return ok, nil
}

// EmailExists tests where a user exists by specified columns.
// Possible columns: query.ColumnEmail, query.ColumnFtcID
func (env Env) EmailExists(email string) (bool, error) {
	var ok bool
	err := env.dbs.Read.Get(&ok, account.StmtEmailExists, email)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (env Env) NameExists(name string) (bool, error) {
	var ok bool
	err := env.dbs.Read.Get(&ok, account.StmtNameExists, name)
	if err != nil {
		return false, err
	}

	return ok, nil
}
