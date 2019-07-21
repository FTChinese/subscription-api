package model

import (
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
)

// FindFtcUser retrieves an FTC account's basic info.
func (env Env) FindFtcUser(ftcId string) (paywall.FtcUser, error) {

	var u paywall.FtcUser
	err := env.db.QueryRow(query.SelectFtcUser, ftcId).Scan(
		&u.UserID,
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

func (env Env) FindStripeCustomer(cusID string) (paywall.FtcUser, error) {
	var u paywall.FtcUser
	err := env.db.QueryRow(query.SelectStripeCustomer, cusID).Scan(
		&u.UserID,
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
