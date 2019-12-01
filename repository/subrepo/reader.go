package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

// FindFtcUser retrieves an FTC account's basic info.
func (env SubEnv) FindFtcUser(ftcId string) (reader.Account, error) {

	var u reader.Account
	err := env.db.Get(
		&u,
		query.SelectFtcUser,
		ftcId,
	)

	if err != nil {
		logger.WithField("trace", "SubEnv.FindFtcUser").Error(err)

		return u, err
	}

	return u, nil
}

func (env SubEnv) FindStripeCustomer(cusID string) (reader.Account, error) {
	var u reader.Account
	err := env.db.QueryRow(query.SelectStripeCustomer, cusID).Scan(
		&u.FtcID,
		&u.UnionID,
		&u.StripeID,
		&u.UserName,
		&u.Email,
	)

	if err != nil {
		logger.WithField("trace", "SubEnv.FindStripeCustomer").Error(err)

		return u, err
	}

	return u, nil
}
