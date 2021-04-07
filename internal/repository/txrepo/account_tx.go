package txrepo

import (
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/jmoiron/sqlx"
)

type AccountTx struct {
	*sqlx.Tx
}

func NewAccountTx(tx *sqlx.Tx) AccountTx {
	return AccountTx{
		Tx: tx,
	}
}

// BaseAccountForStripe loads an ftc account by uuid
func (tx AccountTx) BaseAccountForStripe(ftcID string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := tx.Get(
		&a,
		account.StmtLockBaseAccount,
		ftcID)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

func (tx AccountTx) SavedStripeID(a account.BaseAccount) error {
	_, err := tx.NamedExec(
		account.StmtSetStripeID,
		a,
	)
	if err != nil {
		return err
	}

	return nil
}
