package accounts

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/account"
)

// LoadAddress retrieve user address.
// If returned error is ErrNoRows, it does not mean the user does not exist since we split those data into a separate table the the record might not exist in the new table.
// Simply return zero values since this request is always sent by a logged in user.
func (env Env) LoadAddress(ftcID string) (account.Address, error) {

	var addr account.Address
	err := env.DBs.Read.Get(
		&addr,
		account.StmtLoadAddress,
		ftcID,
	)

	if err != nil {
		// Address is migrated to another table.
		// Data might not moved yet.
		if err == sql.ErrNoRows {
			return account.Address{}, nil
		}
		return addr, err
	}

	return addr, nil
}

// UpdateAddress updates user's physical address.
func (env Env) UpdateAddress(addr account.Address) error {

	_, err := env.DBs.Write.NamedExec(
		account.StmtUpdateAddress,
		addr,
	)

	if err != nil {
		return err
	}

	return nil
}
