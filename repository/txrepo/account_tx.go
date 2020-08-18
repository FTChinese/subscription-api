package txrepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("package", "query")

type AccountTx struct {
	*sqlx.Tx
}

func NewAccountTx(tx *sqlx.Tx) AccountTx {
	return AccountTx{tx}
}

// RetrieveAccount loads an ftc account by uuid
func (tx AccountTx) RetrieveAccount(ftcID string) (reader.Account, error) {
	var account reader.Account
	err := tx.Get(
		&account,
		reader.StmtAccountByFtcID+" FOR UPDATE",
		ftcID)
	if err != nil {
		logger.WithField("trace", "AccountTx.RetrieveAccount").Error(err)
		return reader.Account{}, err
	}

	return account, nil
}

func (tx AccountTx) SavedStripeID(account reader.Account) error {
	_, err := tx.NamedExec(
		reader.StmtSetStripeID,
		account,
	)
	if err != nil {
		logger.WithField("trace", "AccountTx.SaveStripeID").Error(err)
		return err
	}

	return nil
}
