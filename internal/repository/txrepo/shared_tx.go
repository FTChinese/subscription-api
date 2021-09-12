package txrepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/jmoiron/sqlx"
)

type SharedTx struct {
	*sqlx.Tx
}

func NewSharedTx(tx *sqlx.Tx) SharedTx {
	return SharedTx{
		Tx: tx,
	}
}

// CreateMember creates a new membership.
func (tx SharedTx) CreateMember(m reader.Membership) error {

	_, err := tx.NamedExec(
		reader.StmtCreateMember,
		m)

	if err != nil {
		return err
	}

	return nil
}

// RetrieveMember retrieves a user's membership by a compound id, which might be ftc id or union id.
// Use SQL FIND_IN_SET(compoundId, vip_id, vip) to verify it against two columns.
// Returns zero value of membership if not found.
func (tx SharedTx) RetrieveMember(compoundID string) (reader.Membership, error) {
	var m reader.Membership

	err := tx.Get(
		&m,
		reader.StmtLockMember,
		compoundID,
	)

	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	// Treat a non-existing member as a valid value.
	return m.Sync(), nil
}

// UpdateMember updates existing membership.
func (tx SharedTx) UpdateMember(m reader.Membership) error {
	_, err := tx.NamedExec(
		reader.StmtUpdateMember,
		m)

	if err != nil {
		return err
	}

	return nil
}

// DeleteMember deletes a membership.
// This is used both when linking and unlinking.
// When linking IAP to FTC account, all existing membership
// will be deleted and newly merged or created one will
// be inserted.
// When unlinking, the membership is simply deleted, which
// is the correct operation since the membership is granted
// by IAP. You cannot simply remove the apple_subscription_id
// column which will keep the membership on FTC account.
func (tx SharedTx) DeleteMember(id ids.UserIDs) error {
	_, err := tx.NamedExec(reader.StmtDeleteMember, id)

	if err != nil {
		return err
	}

	return nil
}

// SaveInvoice inserts a new invoice to db.
func (tx SharedTx) SaveInvoice(inv invoice.Invoice) error {
	_, err := tx.NamedExec(invoice.StmtCreateInvoice, inv)
	if err != nil {
		return err
	}

	return nil
}
