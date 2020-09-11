package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/client"
)

func (env Env) SaveOrderClient(c client.OrderClient) error {

	_, err := env.db.NamedExec(
		client.StmtInsertOrderClient,
		c)

	if err != nil {
		return err
	}

	return nil
}
