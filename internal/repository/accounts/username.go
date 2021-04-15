package accounts

import "github.com/FTChinese/subscription-api/pkg/account"

func (env Env) UpdateUserName(a account.BaseAccount) error {
	_, err := env.DBs.Write.NamedExec(account.StmtUpdateUserName, a)
	if err != nil {
		return err
	}

	return nil
}
