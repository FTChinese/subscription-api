package model

import (
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/stripepay"
)

// CreateStripeCustomer create a customer under ftc account
// for user with `ftcID`.
func (env Env) CreateStripeCustomer(ftcID string) (string, error) {
	tx, err := env.db.Begin()
	if err != nil {
		logger.WithField("trace", "CreateStripeCustomer").Error(err)
		return "", err
	}

	var u paywall.FtcUser
	err = env.db.QueryRow(query.LockFtcUser, ftcID).Scan(
		&u.UserID,
		&u.UnionID,
		&u.StripeID,
		&u.UserName,
		&u.Email)
	if err != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "CreateStripeCustomer").Error(err)
		return "", err
	}

	if u.StripeID.Valid {
		_ = tx.Rollback()
		return u.StripeID.String, nil
	}

	stripeID, err := stripepay.CreateCustomer(u.Email)
	if err != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "CreateStripeCustomer").Error(err)
		return "", err
	}

	_, err = tx.Exec(query.SaveStripeID, stripeID, ftcID)
	if err != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "CreateStripeCustomer").Error(err)
		return "", err
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("trace", "CreateStripeCustomer").Error(err)
		return "", err
	}

	return stripeID, nil
}
