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
		logger.WithField("trace", "EnvFindUser").Error(err)

		return u, err
	}

	return u, nil
}
