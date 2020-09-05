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
		logger.WithField("trace", "Env.FindAccountByFtcID").Error(err)

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
		logger.WithField("trace", "Env.FindAccountByStripeID").Error(err)

		return u, err
	}

	return u, nil
}
