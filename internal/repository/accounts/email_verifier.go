package accounts

import "github.com/FTChinese/subscription-api/pkg/account"

func (env Env) SaveEmailVerifier(v account.EmailVerifier) error {
	_, err := env.DBs.Write.NamedExec(account.StmtInsertEmailVerifier, v)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) RetrieveEmailVerifier(token string) (account.EmailVerifier, error) {
	var v account.EmailVerifier
	err := env.DBs.Read.Get(&v, account.StmtRetrieveEmailVerifier, token)
	if err != nil {
		return account.EmailVerifier{}, err
	}

	return v, nil
}

func (env Env) EmailVerified(ftcID string) error {
	_, err := env.DBs.Write.Exec(account.StmtEmailVerified, ftcID)

	if err != nil {
		return err
	}

	return nil
}
