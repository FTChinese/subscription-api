package accounts

import (
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/jmoiron/sqlx"
)

// AccountTx performs linking ftc account with wechat in a
// transaction.
type AccountTx struct {
	*sqlx.Tx
}

func NewAccountTx(tx *sqlx.Tx) AccountTx {

	return AccountTx{
		Tx: tx,
	}
}

// CreateAccount create a ftc account.
func (tx AccountTx) CreateAccount(a account.BaseAccount) error {

	_, err := tx.NamedExec(
		account.StmtCreateFtc,
		a)

	if err != nil {
		return err
	}

	return nil
}

// CreateProfile creates profile when creating ftc account.
func (tx AccountTx) CreateProfile(a account.BaseAccount) error {
	_, err := tx.NamedExec(
		account.StmtCreateProfile,
		a)

	if err != nil {
		return err
	}

	return nil
}
