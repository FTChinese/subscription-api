package addons

import (
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/invoice"
)

func (env Env) InvoicesCarriedOver(userID pkg.UserIDs) error {
	_, err := env.dbs.Write.Exec(invoice.StmtCarriedOver, userID.BuildFindInSet())
	if err != nil {
		return err
	}

	return nil
}
