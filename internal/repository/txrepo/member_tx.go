package txrepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/apple"
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
func (tx MemberTx) RetrieveMember(id reader.MemberID) (reader.Membership, error) {
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
	return m.Normalize(), nil
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

	return m.Normalize(), nil
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

// RetrieveOrder loads a previously saved order that is not
// confirmed yet.
func (tx MemberTx) RetrieveOrder(orderID string) (subs.Order, error) {
	var order subs.Order

	err := tx.Get(
		&order,
		subs.StmtLockOrder,
		orderID,
	)

	if err != nil {
		return order, err
	}

	return order, nil
}

func (tx MemberTx) LockOrder(orderID string) (subs.LockedOrder, error) {
	var lo subs.LockedOrder

	err := tx.Get(&lo, subs.StmtLockOrderExpedient, orderID)

	if err != nil {
		return subs.LockedOrder{}, err
	}

	return lo, nil
}

// ConfirmOrder set an order's confirmation time.
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

// CreateMember creates a new membership.
func (tx MemberTx) CreateMember(m reader.Membership) error {
	m = m.Normalize()

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
	m = m.Normalize()

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
func (tx MemberTx) DeleteMember(id reader.MemberID) error {
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

// FindBalanceSources retrieves all orders that has unused portions.
// Used to build upgrade order for alipay and wxpay
func (tx MemberTx) FindBalanceSources(userIDs reader.MemberID) ([]subs.BalanceSource, error) {

	var orders = make([]subs.BalanceSource, 0)

	err := tx.Select(
		&orders,
		subs.StmtBalanceSource,
		userIDs.BuildFindInSet())

	if err != nil {
		return nil, err
	}

	return orders, nil
}

// SaveProratedOrders saved user's current total balance
// the the upgrade plan at this moment.
// It also saves all orders with unused portion to calculate each order's balance.
// Go's SQL does not support batch insert now.
// We use a loop here to insert all record.
// Most users won't have much  valid orders
// at a specific moment, so this should not pose a severe
// performance issue.
func (tx MemberTx) SaveProratedOrders(po []subs.ProratedOrder) error {

	for _, v := range po {
		_, err := tx.NamedExec(
			subs.StmtSaveProratedOrder,
			v)

		if err != nil {
			return err
		}
	}

	return nil
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
