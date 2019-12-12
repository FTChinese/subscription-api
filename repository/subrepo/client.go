package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

func (env SubEnv) SaveOrderClient(c subscription.OrderClient) error {

	_, err := env.db.NamedExec(
		query.BuildInsertClientApp(env.UseSandboxDB()),
		c)

	if err != nil {
		return err
	}

	return nil
}
