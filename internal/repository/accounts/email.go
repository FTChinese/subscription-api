package accounts

import "github.com/FTChinese/subscription-api/pkg/account"

func (env Env) UpdateEmail(a account.BaseAccount) error {
	_, err := env.DBs.Write.NamedExec(account.StmtUpdateEmail, a)
	if err != nil {
		return err
	}

	return nil
}

func (env Env) SaveEmailHistory(a account.BaseAccount) error {
	_, err := env.DBs.Write.NamedExec(account.StmtBackUpEmail, a)

	return err
}
