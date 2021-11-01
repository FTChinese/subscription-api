package accounts

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
)

// Authenticate verifies email+password are correct.
// If the returned error is sql.ErrorNoRow, it indicates the
// email does not exists.
// If no error returned, the AuthResult.PasswordMatched
// field indicates whether the password is correct.
func (env Env) Authenticate(params input.EmailCredentials) (account.AuthResult, error) {
	var r account.AuthResult
	err := env.DBs.Read.Get(&r,
		account.StmtVerifyEmailPassword,
		params.Password,
		params.Email)

	if err != nil {
		return r, err
	}

	return r, nil
}

func (env Env) SignUpCount(params account.SignUpRateParams) (account.SignUpLimit, error) {
	var limit account.SignUpLimit
	err := env.DBs.Read.Get(&limit,
		account.StmtSignUpCount,
		params.StartAt,
		params.EndAt,
		params.IP)

	if err != nil {
		if err == sql.ErrNoRows {
			return limit, nil
		}
		return limit, err
	}

	return limit, nil
}

func (env Env) VerifyIDPassword(params account.IDCredentials) (account.AuthResult, error) {
	var matched bool
	err := env.DBs.Read.Get(
		&matched,
		account.StmtVerifyPassword,
		params.Password,
		params.FtcID)

	if err != nil {
		return account.AuthResult{}, err
	}

	return account.AuthResult{
		UserID:          params.FtcID,
		PasswordMatched: matched,
	}, nil
}

// UpdatePassword updates reader's password.
// This is used both by resetting password if forgotten and updating password after logged in.
func (env Env) UpdatePassword(p account.IDCredentials) error {

	_, err := env.DBs.Write.NamedExec(
		account.StmtUpdatePassword,
		p)

	if err != nil {
		return err
	}

	return nil
}
