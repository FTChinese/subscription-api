package accounts

import "github.com/FTChinese/subscription-api/pkg/account"

// CreateProfile ensures user exists under
// user_db.profile table.
// This is usually used after user performed login.
func (env Env) CreateProfile(a account.BaseAccount) error {
	_, err := env.DBs.Write.NamedExec(
		account.StmtCreateProfile,
		a)
	if err != nil {
		return err
	}

	return nil
}

// LoadProfile shows all of a user's data.
// LoadProfile does not include membership data, which might complicate the query a lot.
func (env Env) LoadProfile(ftcID string) (account.Profile, error) {

	var s account.ProfileSchema

	err := env.DBs.Read.Get(
		&s,
		account.StmtProfile,
		ftcID,
	)

	if err != nil {
		return account.Profile{}, err
	}

	return s.Profile(), nil
}

// UpdateProfile update a user's gender, familyName, givenName, birthday.
func (env Env) UpdateProfile(p account.BaseProfile) error {

	_, err := env.DBs.Write.NamedExec(
		account.StmtUpdateProfile,
		p,
	)

	if err != nil {
		return err
	}

	return nil
}
