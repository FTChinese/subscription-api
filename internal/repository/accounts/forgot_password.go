package accounts

import (
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/account"
)

// SavePwResetSession saves a PwResetSession used to generate
// a password reset email.
func (env Env) SavePwResetSession(s account.PwResetSession) error {

	_, err := env.DBs.Write.NamedExec(
		account.StmtInsertPwResetSession,
		s,
	)

	if err != nil {
		return err
	}

	return nil
}

// PwResetSessionByToken retrieves PwResetSession by token.
func (env Env) PwResetSessionByToken(token string) (account.PwResetSession, error) {
	var session account.PwResetSession
	err := env.DBs.Read.Get(&session, account.StmtPwResetSessionByToken, token)
	if err != nil {
		return account.PwResetSession{}, err
	}

	return session, nil
}

// PwResetSessionByCode retrieves PwResetSession by app code and email.
// Since the code only has 10^6 combinations, chances of duplicate are very high.
// Use email + code to uniquely identify a row.
func (env Env) PwResetSessionByCode(params pkg.AppResetPwSessionParams) (account.PwResetSession, error) {
	var session account.PwResetSession
	err := env.DBs.Read.Get(&session, account.StmtPwResetSessionByCode, params.AppCode, params.Email)
	if err != nil {
		return account.PwResetSession{}, err
	}

	return session, nil
}

// DisablePasswordReset disables a token used.
func (env Env) DisablePasswordReset(t string) error {

	_, err := env.DBs.Write.Exec(account.StmtDisablePwResetToken, t)

	if err != nil {
		return err
	}

	return nil
}
