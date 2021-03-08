package readerrepo

import (
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/invoice"
)

func (env Env) InvoicesCarriedOver(userID pkg.MemberID) error {
	_, err := env.db.Exec(invoice.StmtCarriedOver, userID.BuildFindInSet())
	if err != nil {
		return err
	}

	return nil
}
