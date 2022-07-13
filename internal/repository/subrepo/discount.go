package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/ids"
)

func (env Env) InsertDiscountRedeemed(r ftcpay.DiscountRedeemed) error {
	_, err := env.dbs.Write.NamedExec(
		ftcpay.StmtInsertDiscountRedeemed,
		r)

	return err
}

func (env Env) RetrieveDiscountRedeemed(userIDs ids.UserIDs, discountID string) (ftcpay.DiscountRedeemed, error) {
	var redeemed ftcpay.DiscountRedeemed
	err := env.dbs.Read.Get(
		&redeemed,
		ftcpay.StmtRetrieveDiscountRedeemed,
		userIDs.BuildFindInSet(),
		discountID)

	if err != nil {
		return ftcpay.DiscountRedeemed{}, err
	}

	return redeemed, nil
}
