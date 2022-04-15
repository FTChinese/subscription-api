package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/lib/sq"
	"github.com/FTChinese/subscription-api/pkg/ids"
)

func (env Env) InsertDiscountRedeemed(r ftcpay.DiscountRedeemed) error {
	_, err := env.dbs.Write.NamedExec(
		ftcpay.StmtInsertDiscountRedeemed,
		r)

	return err
}

func (env Env) IsDiscountRedeemed(userIDs ids.UserIDs, discountID string) (bool, error) {
	var ok bool
	err := env.dbs.Read.Get(
		&ok,
		ftcpay.StmtDiscountRedeemed,
		discountID,
		sq.FindInSetValue(userIDs.CollectIDs()))

	if err != nil {
		return false, err
	}

	return ok, nil
}
