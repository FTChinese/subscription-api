package txrepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/jmoiron/sqlx"
)

// StripeTx extends SharedTx.
type StripeTx struct {
	SharedTx
}

func NewStripeTx(tx *sqlx.Tx) StripeTx {
	return StripeTx{
		SharedTx: NewSharedTx(tx),
	}
}

// BaseAccountForStripe loads the base account by stripe customer id.
func (tx StripeTx) BaseAccountForStripe(ftcID string) (account.BaseAccount, error) {
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

func (tx StripeTx) SaveCustomerID(a account.BaseAccount) error {
	_, err := tx.NamedExec(
		stripe.StmtSetCustomerID,
		a,
	)
	if err != nil {
		return err
	}

	return nil
}

// RetrieveStripeMember retrieves membership by stripe subscription id.
// This is actually not used.
func (tx StripeTx) RetrieveStripeMember(subID string) (reader.Membership, error) {
	var m reader.Membership

	err := tx.Get(
		&m,
		reader.StmtLockStripeMember,
		subID)

	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	return m.Sync(), nil
}
