package repository

import (
	"gitlab.com/ftchinese/subscription-api/models/paywall"
)

func (env Env) SaveConfirmationResult(r *paywall.ConfirmationResult) error {
	_, err := env.db.Exec(
		env.query.ConfirmationResult(),
		r.OrderID,
		r.Succeeded,
		r.Failed)

	if err != nil {
		return err
	}

	return nil
}
