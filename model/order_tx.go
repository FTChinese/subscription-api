package model

import (
	"database/sql"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/util"
	"strings"
)

type OrderTx struct {
	tx    *sql.Tx
	query query.Builder
}

// RetrieveMember retrieves a user's membership info by ftc id
// or wechat union id.
func (o OrderTx) RetrieveMember(u paywall.UserID) (paywall.Membership, error) {
	var m paywall.Membership

	// In the ftc_vip table, vip_id might be ftc uuid or wechat
	// id, and vip_id_alias is always wechat id.
	// In future, the table will be refactor with two extra
	// columns ftc_user_id dedicated to ftc uuid and wx_union_id
	// dedicated for wechat union id. The vip_id column will be
	// use only as a unique constraint on these two columns.
	err := o.tx.QueryRow(
		o.query.SelectMemberLock(),
		u.CompoundID,
		u.UnionID,
	).Scan(
		&m.CompoundID,
		&m.UnionID,
		&m.Tier,
		&m.Cycle,
		&m.ExpireDate)

	if err != nil {
		logger.WithField("trace", "OrderTx.RetrieveMember").Error(err)

		return m, err
	}

	// Treat a non-existing member as a valid value.
	return m, nil
}

// SaveOrder saves an order to db.
func (o OrderTx) SaveOrder(s paywall.Subscription, c util.ClientApp) error {

	_, err := o.tx.Exec(
		o.query.InsertSubs(),
		s.OrderID,
		s.CompoundID,
		s.FtcID,
		s.UnionID,
		s.ListPrice,
		s.NetPrice,
		s.TierToBuy,
		s.BillingCycle,
		s.CycleCount,
		s.ExtraDays,
		s.Kind,
		s.UpgradeBalance,
		s.PaymentMethod,
		s.WxAppID,
		c.ClientType,
		c.Version,
		c.UserIP,
		c.UserAgent)

	if err != nil {
		logger.WithField("trace", "OrderTx.SaveSubscription").Error(err)
		return err
	}

	return nil
}

// FindUnusedOrders retrieves all orders that has unused portions.
func (o OrderTx) FindUnusedOrders(u paywall.UserID) ([]paywall.UnusedOrder, error) {
	rows, err := o.tx.Query(
		o.query.UnusedOrders(),
		u.CompoundID,
		u.UnionID)
	if err != nil {
		logger.WithField("trace", "OrderTx.FindUnusedOrders").Error(err)
		return nil, err
	}
	defer rows.Close()

	orders := make([]paywall.UnusedOrder, 0)
	for rows.Next() {
		var o paywall.UnusedOrder

		err := rows.Scan(
			&o.ID,
			&o.NetPrice,
			&o.StartDate,
			&o.EndDate)

		if err != nil {
			logger.WithField("trace", "OrderTx.FindUnusedOrders").Error(err)
			return nil, err
		}

		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		logger.WithField("trace", "OrderTx.FindUnusedOrders").Error(err)
		return nil, err
	}

	return orders, nil
}

// BuildUpgradeOrder tries to find out unused orders
// and build a Subscription based on those orders.
func (o OrderTx) BuildUpgradeOrder(user paywall.UserID, plan paywall.Plan) (paywall.Subscription, error) {
	orders, err := o.FindUnusedOrders(user)
	if err != nil {
		return paywall.Subscription{}, err
	}

	up := paywall.NewUpgradePlan(plan).
		SetBalance(orders).
		CalculatePayable()

	subs, err := paywall.NewSubsUpgrade(user, up)
	if err != nil {
		logger.WithField("trace", "OrderTx.BuildUpgradeOrder").Error(err)
		return subs, err
	}

	return subs, nil
}

// SaveUpgradeSource saves unused orders to db.
func (o OrderTx) SaveUpgradeSource(targetID string, srcIDs []string) error {
	id := strings.Join(srcIDs, ",")

	_, err := o.tx.Exec(o.query.InsertUpgradeSource(), targetID, id)

	if err != nil {
		logger.WithField("trace", "OrderTx.SaveUpgradeSource").Error(err)
		return err
	}

	return nil
}

func (o OrderTx) rollback() error {
	return o.tx.Rollback()
}

func (o OrderTx) commit() error {
	return o.tx.Commit()
}
