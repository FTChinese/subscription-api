package accounts

import (
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/client"
)

func (env Env) SaveClient(c account.ClientFootprint) error {
	_, err := env.DBs.Write.NamedExec(client.StmtInsertClient, c)

	if err != nil {
		return err
	}

	return nil
}
