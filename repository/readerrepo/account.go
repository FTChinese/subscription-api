package readerrepo

import "github.com/FTChinese/subscription-api/pkg/reader"

func (env Env) AccountByFtcID(id string) (reader.Account, error) {
	var u reader.Account
	err := env.db.Get(
		&u,
		reader.StmtAccountByFtcID,
		id,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

func (env Env) AccountByStripeID(cusID string) (reader.Account, error) {
	var u reader.Account
	err := env.db.Get(&u,
		reader.StmtAccountByStripeID,
		cusID)

	if err != nil {
		return u, err
	}

	return u, nil
}

func (env Env) SandboxUserExists(ftcID string) (bool, error) {
	var found bool
	err := env.db.Get(&found, reader.StmtSandboxExists, ftcID)
	if err != nil {
		return false, err
	}

	return found, nil
}
