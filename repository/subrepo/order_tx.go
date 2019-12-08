package subrepo

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
	tx      *sqlx.Tx
	sandbox bool // Indicates whether we should use the sandbox DB.
}

// RetrieveMember retrieves a user's membership info by ftc id
// or wechat union id.
// Returns zero value of membership if not found.
func (otx OrderTx) RetrieveMember(id reader.MemberID) (subscription.Membership, error) {
	var m subscription.Membership

	err := otx.tx.Get(
		&m,
		query.BuildSelectMembership(otx.sandbox, true),
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
func (otx OrderTx) SaveOrder(order subscription.Order) error {

	_, err := otx.tx.NamedExec(
		query.BuildInsertOrder(otx.sandbox),
		order)

	if err != nil {
		logger.WithField("trace", "OrderTx.SaveSubscription").Error(err)
		return err
	}

	return nil
}

// RetrieveOrder loads a previously saved order that is not
// confirmed yet.
func (otx OrderTx) RetrieveOrder(orderID string) (subscription.Order, error) {
	var order subscription.Order

	err := otx.tx.Get(
		&order,
		query.BuildSelectOrder(otx.sandbox),
		orderID,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.RetrieveOrder").Error(err)

		return order, err
	}

	return order, nil
}

// UpdateConfirmedOrder set an order's confirmation time.
func (otx OrderTx) UpdateConfirmedOrder(order subscription.Order) error {
	_, err := otx.tx.NamedExec(
		query.BuildConfirmOrder(otx.sandbox),
		order,
	)

	if err != nil {
		logger.WithField("trace", "OrderTx.UpdateConfirmedOrder").Error(err)

		return err
	}

	return nil
}

func (otx OrderTx) CreateMember(m subscription.Membership) error {
	m.Normalize()

	_, err := otx.tx.NamedExec(
		query.BuildInsertMembership(otx.sandbox),
		m,
	)

	if err != nil {
		logger.WithField("trace", "OrderTx.CreateMember").Error(err)
		return err
	}

	return nil
}

func (otx OrderTx) UpdateMember(m subscription.Membership) error {
	m.Normalize()

	_, err := otx.tx.NamedExec(
		query.BuildUpdateMembership(otx.sandbox),
		m)

	if err != nil {
		logger.WithField("trace", "OrderTx.UpdateMembership").Error(err)
		return err
	}

	return nil
}

// FindBalanceSources retrieves all orders that has unused portions.
// Used to build upgrade order for alipay and wxpay
// TODO: change sql statement.
func (otx OrderTx) FindBalanceSources(accountID reader.MemberID) ([]subscription.ProratedOrder, error) {

	var orders = make([]subscription.ProratedOrder, 0)

	err := otx.tx.Select(
		&orders,
		query.BuildSelectBalanceSource(otx.sandbox),
		accountID.CompoundID,
		accountID.UnionID)

	if err != nil {
		logger.WithField("trace", "OrderTx.FindBalanceSources").Error(err)
		return nil, err
	}

	return orders, nil
}

// SaveProratedOrders saves all orders with unused portion to calculate each order's balance.
// Go's SQL does not support batch insert now.
// We use a loop here to insert all record.
// Most users won't have much  valid orders
// at a specific moment, so this should not pose a severe
// performance issue.
func (otx OrderTx) SaveProratedOrders(p []subscription.ProratedOrderSchema) error {
	for _, v := range p {
		_, err := otx.tx.NamedExec(
			query.BuildInsertProration(otx.sandbox),
			v)

		if err != nil {
			return err
		}
	}

	return nil
}

// SaveUpgradeIntent saved user's current total balance
// the the upgrade plan at this moment.
// TODO: check sql
func (otx OrderTx) SaveUpgradeIntent(up subscription.UpgradeSchema) error {

	_, err := otx.tx.NamedExec(
		query.BuildInsertUpgradePlan(otx.sandbox),
		up)

	if err != nil {
		return err
	}

	return nil
}

// ConfirmUpgrade set an upgrade's confirmation time.
func (otx OrderTx) ConfirmUpgrade(upgradeID string) error {
	_, err := otx.tx.Exec(
		query.BuildProrationUsed(otx.sandbox),
		upgradeID,
	)
	if err != nil {
		logger.WithField("trace", "OrderTx.ConfirmUpgrade").Error(err)
		return err
	}

	return nil
}

// -------------
// The following are used by gift card

func (otx OrderTx) ActivateGiftCard(code string) error {
	_, err := otx.tx.Exec(
		query.BuildActivateGiftCard(otx.sandbox),
		code)

	if err != nil {
		return err
	}

	return nil
}

func (otx OrderTx) Rollback() error {
	return otx.tx.Rollback()
}

func (otx OrderTx) Commit() error {
	return otx.tx.Commit()
}
