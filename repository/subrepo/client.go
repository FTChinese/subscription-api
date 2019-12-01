package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

func (env SubEnv) SaveOrderClient(orderID string, app util.ClientApp) error {
	data := struct {
		OrderID string `db:"order_id"`
		util.ClientApp
	}{
		orderID,
		app,
	}

	_, err := env.db.NamedExec(
		query.BuildInsertClientApp(env.Sandbox),
		data)

	if err != nil {
		return err
	}

	return nil
}
