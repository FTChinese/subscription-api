package txrepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
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

// RetrieveAccount loads an ftc account by uuid
func (tx AccountTx) RetrieveAccount(ftcID string) (reader.FtcAccount, error) {
	var account reader.FtcAccount
	err := tx.Get(
		&account,
		reader.StmtAccountByFtcID+" FOR UPDATE",
		ftcID)
	if err != nil {
		return reader.FtcAccount{}, err
	}

	return account, nil
}

func (tx AccountTx) SavedStripeID(account reader.FtcAccount) error {
	_, err := tx.NamedExec(
		reader.StmtSetStripeID,
		account,
	)
	if err != nil {
		return err
	}

	return nil
}
