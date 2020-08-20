package txrepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/redeem"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/jmoiron/sqlx"
)

// MemberTx check a user's member status and create an order
// if allowed.
type MemberTx struct {
	*sqlx.Tx
	dbName config.SubsDB
}

func NewMemberTx(tx *sqlx.Tx, cfg config.BuildConfig) MemberTx {
	return MemberTx{
		Tx:     tx,
		dbName: cfg.GetSubsDB(),
	}
}

// RetrieveMember retrieves a user's membership by ftc id
// or wechat union id.
// Returns zero value of membership if not found.
func (tx MemberTx) RetrieveMember(id reader.MemberID) (subs.Membership, error) {
	var m subs.Membership

	err := tx.Get(
		&m,
		subs.StmtLockMember(tx.dbName),
		id.CompoundID,
	)

	if err != nil && err != sql.ErrNoRows {
		logger.WithField("trace", "MemberTx.RetrieveMember").Error(err)

		return m, err
	}

	// Normalize legacy columns
	m.Normalize()

	// Treat a non-existing member as a valid value.
	return m, nil
}

// RetrieveAppleMember selects membership by apple original transaction id.
// // NOTE: sql.ErrNoRows are ignored. The returned
//// Membership might be a zero value.
func (tx MemberTx) RetrieveAppleMember(transactionID string) (subs.Membership, error) {
	var m subs.Membership

	err := tx.Get(
		&m,
		subs.StmtAppleMember(tx.dbName),
		transactionID)

	if err != nil && err != sql.ErrNoRows {
		logger.WithField("trace", "MemberTx.RetrieveAppleMember").Error(err)

		return m, err
	}

	m.Normalize()

	return m, nil
}

// SaveOrder saves an order to db.
// This is only limited to alipay and wechat pay.
// Stripe pay does not generate any orders on our side.
func (tx MemberTx) SaveOrder(order subs.Order) error {

	_, err := tx.NamedExec(
		subs.StmtCreateOrder(tx.dbName),
		order)

	if err != nil {
		logger.WithField("trace", "MemberTx.SaveSubscription").Error(err)
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
		subs.StmtOrder(tx.dbName),
		orderID,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.RetrieveOrder").Error(err)

		return order, err
	}

	return order, nil
}

// ConfirmOrder set an order's confirmation time.
func (tx MemberTx) ConfirmOrder(order subs.Order) error {
	_, err := tx.NamedExec(
		subs.StmtConfirmOrder(tx.dbName),
		order,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.ConfirmOrder").Error(err)

		return err
	}

	return nil
}

// CreateMember creates a new membership.
func (tx MemberTx) CreateMember(m subs.Membership) error {
	m.Normalize()

	_, err := tx.NamedExec(
		subs.StmtCreateMember(tx.dbName),
		m,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.CreateMember").Error(err)
		return err
	}

	return nil
}

// UpdateMember updates existing membership.
func (tx MemberTx) UpdateMember(m subs.Membership) error {
	m.Normalize()

	_, err := tx.NamedExec(
		subs.StmtUpdateMember(tx.dbName),
		m)

	if err != nil {
		logger.WithField("trace", "MemberTx.UpdateMembership").Error(err)
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
		subs.StmtDeleteMember(tx.dbName),
		id)

	if err != nil {
		logger.WithField("trace", "MembershipTx.DeleteMember").Error(err)

		return err
	}

	return nil
}

// FindBalanceSources retrieves all orders that has unused portions.
// Used to build upgrade order for alipay and wxpay
func (tx MemberTx) FindBalanceSources(userIDs reader.MemberID) ([]subs.ProratedOrder, error) {

	var orders = make([]subs.ProratedOrder, 0)

	err := tx.Select(
		&orders,
		subs.StmtBalanceSource(tx.dbName),
		userIDs.BuildFindInSet())

	if err != nil {
		logger.WithField("trace", "MemberTx.FindBalanceSources").Error(err)
		return nil, err
	}

	return orders, nil
}

// SaveUpgradeSchema saved user's current total balance
// the the upgrade plan at this moment.
// It also saves all orders with unused portion to calculate each order's balance.
// Go's SQL does not support batch insert now.
// We use a loop here to insert all record.
// Most users won't have much  valid orders
// at a specific moment, so this should not pose a severe
// performance issue.
func (tx MemberTx) SaveUpgradeSchema(up subs.UpgradeSchema) error {
	_, err := tx.NamedExec(
		subs.StmtSaveUpgradeBalance(tx.dbName),
		up)

	if err != nil {
		logger.WithField("trace", "MemberTx.SaveUpgradeSchema").Error(err)
		return err
	}

	for _, v := range up.Sources {
		_, err := tx.NamedExec(
			subs.StmtSaveProratedOrder(tx.dbName),
			v)

		if err != nil {
			logger.WithField("trace", "MemberTx.SaveUpgradeSchema").Error(err)
			return err
		}
	}

	return nil
}

// ProratedOrdersUsed set the consumed time on all the
// prorated order for an upgrade operation.
func (tx MemberTx) ProratedOrdersUsed(upgradeID string) error {
	_, err := tx.Exec(
		subs.StmtProratedOrdersUsed(tx.dbName),
		upgradeID,
	)
	if err != nil {
		logger.WithField("trace", "MemberTx.ProratedOrdersUsed").Error(err)
		return err
	}

	return nil
}

// -------------
// The following are used by gift card

func (tx MemberTx) ActivateGiftCard(code string) error {
	_, err := tx.Exec(
		redeem.StmtActivateGiftCard(tx.dbName),
		code)

	if err != nil {
		return err
	}

	return nil
}
