package readerrepo

import (
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env Env) InvoicesCarriedOver(userID reader.MemberID) error {
	_, err := env.db.Exec(invoice.StmtCarriedOver, userID.BuildFindInSet())
	if err != nil {
		return err
	}

	return nil
}
