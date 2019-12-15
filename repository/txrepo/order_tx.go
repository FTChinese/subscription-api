package txrepo

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

// OrderTx check a user's member status and create an order
// if allowed.
type OrderTx struct {
	*sqlx.Tx
	sandbox bool // Indicates whether we should use the sandbox DB.
}

func NewOrderTx(tx *sqlx.Tx, sandbox bool) OrderTx {
	return OrderTx{
		Tx:      tx,
		sandbox: sandbox,
	}
}

// RetrieveMember retrieves a user's membership info by ftc id
// or wechat union id.
// Returns zero value of membership if not found.
func (tx OrderTx) RetrieveMember(id reader.MemberID) (subscription.Membership, error) {
	var m subscription.Membership

	err := tx.Get(
		&m,
		query.BuildSelectMembership(tx.sandbox, true),
		id.CompoundID,
	)

	if err != nil && err != sql.ErrNoRows {
		logger.WithField("trace", "OrderTx.RetrieveMember").Error(err)

		return m, err
	}

	// Normalize legacy columns
	m.Normalize()

	// Treat a non-existing member as a valid value.
	return m, nil
}

// SaveOrder saves an order to db.
// This is only limited to alipay and wechat pay.
// Stripe pay does not generate any orders on our side.
func (tx OrderTx) SaveOrder(order subscription.Order) error {

	_, err := tx.NamedExec(
		query.BuildInsertOrder(tx.sandbox),
		order)

	if err != nil {
		logger.WithField("trace", "OrderTx.SaveSubscription").Error(err)
		return err
	}

	return nil
}

// RetrieveOrder loads a previously saved order that is not
// confirmed yet.
func (tx OrderTx) RetrieveOrder(orderID string) (subscription.Order, error) {
	var order subscription.Order

	err := tx.Get(
		&order,
		query.BuildSelectOrder(tx.sandbox),
		orderID,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.RetrieveOrder").Error(err)

		return order, err
	}

	return order, nil
}

// UpdateConfirmedOrder set an order's confirmation time.
func (tx OrderTx) UpdateConfirmedOrder(order subscription.Order) error {
	_, err := tx.NamedExec(
		query.BuildConfirmOrder(tx.sandbox),
		order,
	)

	if err != nil {
		logger.WithField("trace", "OrderTx.UpdateConfirmedOrder").Error(err)

		return err
	}

	return nil
}

// CreateMember creates a new membership.
func (tx OrderTx) CreateMember(m subscription.Membership) error {
	m.Normalize()

	_, err := tx.NamedExec(
		query.BuildInsertMembership(tx.sandbox),
		m,
	)

	if err != nil {
		logger.WithField("trace", "OrderTx.CreateMember").Error(err)
		return err
	}

	return nil
}

// UpdateMember updates existing membership.
func (tx OrderTx) UpdateMember(m subscription.Membership) error {
	m.Normalize()

	_, err := tx.NamedExec(
		query.BuildUpdateMembership(tx.sandbox),
		m)

	if err != nil {
		logger.WithField("trace", "OrderTx.UpdateMembership").Error(err)
		return err
	}

	return nil
}

// FindBalanceSources retrieves all orders that has unused portions.
// Used to build upgrade order for alipay and wxpay
func (tx OrderTx) FindBalanceSources(accountID reader.MemberID) ([]subscription.ProratedOrder, error) {

	var orders = make([]subscription.ProratedOrder, 0)

	err := tx.Select(
		&orders,
		query.BuildSelectBalanceSource(tx.sandbox),
		accountID.CompoundID,
		accountID.UnionID)

	if err != nil {
		logger.WithField("trace", "OrderTx.FindBalanceSources").Error(err)
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
func (tx OrderTx) SaveUpgradeSchema(up subscription.UpgradeSchema) error {
	_, err := tx.NamedExec(
		query.BuildInsertUpgradeBalance(tx.sandbox),
		up)

	if err != nil {
		logger.WithField("trace", "OrderTx.SaveUpgradeSchema").Error(err)
		return err
	}

	for _, v := range up.Sources {
		_, err := tx.NamedExec(
			query.BuildInsertProration(tx.sandbox),
			v)

		if err != nil {
			logger.WithField("trace", "OrderTx.SaveUpgradeSchema").Error(err)
			return err
		}
	}

	return nil
}

// ProratedOrdersUsed set the consumed time on all the
// prorated order for an upgrade operation.
func (tx OrderTx) ProratedOrdersUsed(upgradeID string) error {
	_, err := tx.Exec(
		query.BuildProrationUsed(tx.sandbox),
		upgradeID,
	)
	if err != nil {
		logger.WithField("trace", "OrderTx.ProratedOrdersUsed").Error(err)
		return err
	}

	return nil
}

// -------------
// The following are used by gift card

func (tx OrderTx) ActivateGiftCard(code string) error {
	_, err := tx.Exec(
		query.BuildActivateGiftCard(tx.sandbox),
		code)

	if err != nil {
		return err
	}

	return nil
}
