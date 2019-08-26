package repository

import (
	"gitlab.com/ftchinese/subscription-api/models/query"
	"gitlab.com/ftchinese/subscription-api/models/reader"
)

// FindFtcUser retrieves an FTC account's basic info.
func (env Env) FindFtcUser(ftcId string) (reader.Account, error) {

	var u reader.Account
	err := env.db.Get(
		&u,
		query.SelectFtcUser,
		ftcId,
	)

	if err != nil {
		logger.WithField("trace", "Env.FindFtcUser").Error(err)

		return u, err
	}

	return u, nil
}

func (env Env) FindStripeCustomer(cusID string) (reader.Account, error) {
	var u reader.Account
	err := env.db.QueryRow(query.SelectStripeCustomer, cusID).Scan(
		&u.FtcID,
		&u.UnionID,
		&u.StripeID,
		&u.UserName,
		&u.Email,
	)

	if err != nil {
		logger.WithField("trace", "Env.FindStripeCustomer").Error(err)

		return u, err
	}

	return u, nil
}
