package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

func (env SubEnv) SaveConfirmationResult(r subscription.ConfirmationResult) error {
	_, err := env.db.NamedExec(
		query.InsertConfirmationResult,
		r)

	if err != nil {
		return err
	}

	return nil
}
