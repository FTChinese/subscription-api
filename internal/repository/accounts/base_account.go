package accounts

import (
	"github.com/FTChinese/subscription-api/pkg/account"
)

// CreateAccount creates a new new FTC account.
func (env Env) CreateAccount(a account.BaseAccount) error {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.beginAccountTx()
	if err != nil {
		sugar.Error(err)
		return err
	}

	err = tx.CreateAccount(a)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return err
	}

	err = tx.CreateProfile(a)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

func (env Env) BaseAccountByEmail(email string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.DBs.Read.Get(&a, account.StmtBaseAccountByEmail, email)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

func (env Env) BaseAccountByMobile(mobile string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.DBs.Read.Get(&a, account.StmtBaseAccountByMobile, mobile)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}
