package txrepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/jmoiron/sqlx"
)

type IAPTx struct {
	SharedTx
}

func NewIAPTx(tx *sqlx.Tx) IAPTx {
	return IAPTx{
		SharedTx: NewSharedTx(tx),
	}
}

// RetrieveAppleMember selects membership by apple original transaction id.
// // NOTE: sql.ErrNoRows are ignored. The returned
//// Membership might be a zero value.
func (tx IAPTx) RetrieveAppleMember(transactionID string) (reader.Membership, error) {
	var m reader.Membership

	err := tx.Get(
		&m,
		reader.StmtLockAppleMember,
		transactionID)

	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	return m.Sync(), nil
}

// RetrieveAppleSubs loads apple subscription when linking IAP to ftc account,
// or severing the link.
func (tx SharedTx) RetrieveAppleSubs(origTxID string) (apple.Subscription, error) {
	var s apple.Subscription
	err := tx.Get(&s, apple.StmtLockSubs, origTxID)

	if err != nil {
		return apple.Subscription{}, err
	}

	return s, nil
}

// LinkAppleSubs set ftc_user_id for the specified original_transaction_id.
func (tx SharedTx) LinkAppleSubs(link apple.LinkInput) error {
	_, err := tx.NamedExec(apple.StmtLinkSubs, link)
	if err != nil {
		return err
	}

	return nil
}

func (tx SharedTx) UnlinkAppleSubs(link apple.LinkInput) error {
	_, err := tx.NamedExec(apple.StmtUnlinkSubs, link)
	if err != nil {
		return err
	}

	return nil
}
