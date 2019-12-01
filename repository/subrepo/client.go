package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/util"
)

func (env Env) SaveOrderClient(orderID string, app util.ClientApp) error {
	data := struct {
		OrderID string `db:"order_id"`
		util.ClientApp
	}{
		orderID,
		app,
	}

	_, err := env.db.NamedExec(
		env.query.InsertClientApp(),
		data)

	if err != nil {
		return err
	}

	return nil
}
