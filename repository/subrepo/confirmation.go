package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/subscription"
)

func (env Env) SaveConfirmationResult(r *subscription.ConfirmationResult) error {
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
