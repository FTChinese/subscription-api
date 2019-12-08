package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"time"
)

func (env SubEnv) PreviewUpgrade(id reader.MemberID) error {

	tx, err := env.BeginOrderTx()
	if err != nil {
		return err
	}

	member, err := tx.RetrieveMember(id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := member.PermitAliWxUpgrade(); err != nil {
		_ = tx.Rollback()
		return err
	}

	orders, err := tx.FindBalanceSources(id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	wallet := subscription.NewWallet(orders, time.Now())
}