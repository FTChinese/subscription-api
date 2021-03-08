package txrepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/redeem"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/jmoiron/sqlx"
)

// MemberTx check a user's member status and create an order
// if allowed.
type MemberTx struct {
	*sqlx.Tx
}

func NewMemberTx(tx *sqlx.Tx) MemberTx {
	return MemberTx{
		Tx: tx,
	}
}

// RetrieveMember retrieves a user's membership by a compound id, which might be ftc id or union id.
// Use SQL FIND_IN_SET(compoundId, vip_id, vip) to verify it against two columns.
// Returns zero value of membership if not found.
func (tx MemberTx) RetrieveMember(id pkg.UserIDs) (reader.Membership, error) {
	var m reader.Membership

	err := tx.Get(
		&m,
		reader.StmtGetLockMember,
		id.BuildFindInSet(),
	)

	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	// Treat a non-existing member as a valid value.
	return m.Sync(), nil
}

// RetrieveAppleMember selects membership by apple original transaction id.
// // NOTE: sql.ErrNoRows are ignored. The returned
//// Membership might be a zero value.
func (tx MemberTx) RetrieveAppleMember(transactionID string) (reader.Membership, error) {
	var m reader.Membership

	err := tx.Get(
		&m,
		reader.StmtGetLockAppleMember,
		transactionID)

	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	return m.Sync(), nil
}

func (tx MemberTx) RetrieveStripeMember(subID string) (reader.Membership, error) {
	var m reader.Membership

	err := tx.Get(
		&m,
		reader.StmtGetLockStripeMember,
		subID)

	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	return m.Sync(), nil
}

// SaveOrder saves an order to db.
// This is only limited to alipay and wechat pay.
// Stripe pay does not generate any orders on our side.
func (tx MemberTx) SaveOrder(order subs.Order) error {

	_, err := tx.NamedExec(
		subs.StmtInsertOrder,
		order)

	if err != nil {
		return err
	}

	return nil
}

func (tx MemberTx) LockOrder(orderID string) (subs.LockedOrder, error) {
	var lo subs.LockedOrder

	err := tx.Get(&lo, subs.StmtLockOrder, orderID)

	if err != nil {
		return subs.LockedOrder{}, err
	}

	return lo, nil
}

// ConfirmOrder set an order's confirmation time and the purchased period.
func (tx MemberTx) ConfirmOrder(order subs.Order) error {
	_, err := tx.NamedExec(
		subs.StmtConfirmOrder,
		order,
	)

	if err != nil {
		return err
	}

	return nil
}

// SaveInvoice inserts a new invoice to db.
func (tx MemberTx) SaveInvoice(inv invoice.Invoice) error {
	_, err := tx.NamedExec(invoice.StmtCreateInvoice, inv)
	if err != nil {
		return err
	}

	return nil
}

// AddOnExistsForOrder checks if the specified order already created an invoice.
func (tx MemberTx) AddOnExistsForOrder(orderID string) (bool, error) {
	var ok bool
	err := tx.Get(&ok, invoice.StmtAddOnExistsForOrder, orderID)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (tx MemberTx) AddOnInvoices(ids pkg.UserIDs) ([]invoice.Invoice, error) {
	var inv []invoice.Invoice
	err := tx.Select(&inv, invoice.StmtListAddOnInvoiceLock, ids.BuildFindInSet())
	if err != nil {
		return nil, err
	}

	return inv, nil
}

func (tx MemberTx) AddOnInvoiceConsumed(inv invoice.Invoice) error {
	_, err := tx.NamedExec(invoice.StmtAddOnInvoiceConsumed, inv)
	if err != nil {
		return err
	}

	return nil
}

// CreateMember creates a new membership.
func (tx MemberTx) CreateMember(m reader.Membership) error {
	_, err := tx.NamedExec(
		reader.StmtCreateMember,
		m,
	)

	if err != nil {
		return err
	}

	return nil
}

// UpdateMember updates existing membership.
func (tx MemberTx) UpdateMember(m reader.Membership) error {
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
func (tx MemberTx) DeleteMember(id pkg.UserIDs) error {
	_, err := tx.NamedExec(
		reader.StmtDeleteMember,
		id)

	if err != nil {
		return err
	}

	return nil
}

func (tx MemberTx) RetrieveAppleSubs(origTxID string) (apple.Subscription, error) {
	var s apple.Subscription
	err := tx.Get(&s, apple.StmtLockSubs, origTxID)

	if err != nil {
		return apple.Subscription{}, err
	}

	return s, nil
}

func (tx MemberTx) LinkAppleSubs(link apple.LinkInput) error {
	_, err := tx.NamedExec(apple.StmtLinkSubs, link)
	if err != nil {
		return err
	}

	return nil
}

func (tx MemberTx) UnlinkAppleSubs(link apple.LinkInput) error {
	_, err := tx.NamedExec(apple.StmtUnlinkSubs, link)
	if err != nil {
		return err
	}

	return nil
}

// -------------
// The following are used by gift card

func (tx MemberTx) ActivateGiftCard(code string) error {
	_, err := tx.Exec(
		redeem.StmtActivateGiftCard,
		code)

	if err != nil {
		return err
	}

	return nil
}
