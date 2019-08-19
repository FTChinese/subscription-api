package repository

import (
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/query"
)

// FindFtcUser retrieves an FTC account's basic info.
func (env Env) FindFtcUser(ftcId string) (paywall.Account, error) {

	var u paywall.Account
	err := env.db.QueryRow(query.SelectFtcUser, ftcId).Scan(
		&u.FtcID,
		&u.UnionID,
		&u.StripeID,
		&u.UserName,
		&u.Email,
	)

	if err != nil {
		logger.WithField("trace", "Env.FindFtcUser").Error(err)

		return u, err
	}

	return u, nil
}

func (env Env) FindStripeCustomer(cusID string) (paywall.Account, error) {
	var u paywall.Account
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
