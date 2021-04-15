package accounts

import "github.com/FTChinese/subscription-api/pkg/account"

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
