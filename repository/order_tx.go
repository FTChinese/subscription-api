package repository

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/query"
	"gitlab.com/ftchinese/subscription-api/models/reader"
)

// OrderTx check a user's member status and create an order
// if allowed.
type OrderTx struct {
	tx    *sqlx.Tx
	live  bool
	query query.Builder
}

// RetrieveMember retrieves a user's membership info by ftc id
// or wechat union id.
// Returns zero value of membership if not found.
func (otx OrderTx) RetrieveMember(id reader.MemberID) (paywall.Membership, error) {
	var m paywall.Membership

	err := otx.tx.Get(
		&m,
		otx.query.SelectMemberLock(id.MemberColumn()),
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
func (otx OrderTx) SaveOrder(order paywall.Order) error {

	_, err := otx.tx.NamedExec(
		otx.query.InsertOrder(),
		order)

	if err != nil {
		logger.WithField("trace", "OrderTx.SaveSubscription").Error(err)
		return err
	}

	return nil
}

// RetrieveOrder loads a previously saved order that is not
// confirmed yet.
func (otx OrderTx) RetrieveOrder(orderID string) (paywall.Order, error) {
	var order paywall.Order

	err := otx.tx.Get(
		&order,
		otx.query.SelectSubsLock(),
		orderID,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.RetrieveOrder").Error(err)

		return order, err
	}

	return order, nil
}

// ConfirmOrder set an order's confirmation time.
func (otx OrderTx) ConfirmOrder(order paywall.Order) error {
	_, err := otx.tx.NamedExec(
		otx.query.ConfirmOrder(),
		order,
	)

	if err != nil {
		logger.WithField("trace", "OrderTx.ConfirmOrder").Error(err)

		return err
	}

	return nil
}

func (otx OrderTx) CreateMember(m paywall.Membership) error {
	m.Normalize()

	_, err := otx.tx.NamedExec(
		otx.query.InsertMember(),
		m,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.CreateMember").Error(err)
		return err
	}

	return nil
}

func (otx OrderTx) UpdateMember(m paywall.Membership) error {
	m.Normalize()

	_, err := otx.tx.NamedExec(
		otx.query.UpdateMember(m.MemberColumn()),
		m)

	if err != nil {
		logger.WithField("trace", "OrderTx.UpdateMembership").Error(err)
		return err
	}

	return nil
}

// FindBalanceSources retrieves all orders that has unused portions.
// Used to build upgrade order for alipay and wxpay
func (otx OrderTx) FindBalanceSources(accountID reader.MemberID) ([]paywall.ProrationSource, error) {

	var sources = []paywall.ProrationSource{}

	err := otx.tx.Select(
		&sources,
		otx.query.SelectProrationSource(),
		accountID.CompoundID,
		accountID.UnionID)

	if err != nil {
		logger.WithField("trace", "OrderTx.FindBalanceSources").Error(err)
		return nil, err
	}

	return sources, nil
}

// SaveProration saves all orders with unused portion to calculate each order's balance.
// Go's SQL does not support batch insert now.
// We use a loop here to insert all record.
// Most users won't have much  valid orders
// at a specific moment, so this should not pose a severe
// performance issue.
func (otx OrderTx) SaveProration(p []paywall.ProrationSource) error {
	for _, v := range p {
		_, err := otx.tx.NamedExec(
			otx.query.InsertProration(),
			v)

		if err != nil {
			return err
		}
	}

	return nil
}

// SaveUpgradePlan saved user's current total balance
// the the upgrade plan at this moment.
func (otx OrderTx) SaveUpgradePlan(up paywall.UpgradePlan) error {

	var data = struct {
		paywall.UpgradePlan
		plan.Plan
	}{
		UpgradePlan: up,
		Plan:        up.Plan,
	}
	_, err := otx.tx.NamedExec(
		otx.query.InsertUpgradePlan(),
		data)

	if err != nil {
		return err
	}

	return nil
}

// ConfirmUpgrade set an upgrade's confirmation time.
func (otx OrderTx) ConfirmUpgrade(upgradeID string) error {
	_, err := otx.tx.Exec(
		otx.query.ProrationConsumed(),
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
	_, err := otx.tx.Exec(otx.query.ActivateGiftCard(), code)

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
