package repository

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/query"
	"strings"
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
func (otx OrderTx) RetrieveMember(id paywall.AccountID) (paywall.Membership, error) {
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
func (otx OrderTx) SaveOrder(s paywall.Subscription, c util.ClientApp) error {

	// Should we move client data to a separate table?
	data := struct {
		paywall.Subscription
		util.ClientApp
	}{
		Subscription: s,
		ClientApp:    c,
	}
	_, err := otx.tx.NamedExec(
		otx.query.InsertSubs(),
		data)

	if err != nil {
		logger.WithField("trace", "OrderTx.SaveSubscription").Error(err)
		return err
	}

	return nil
}

// RetrieveOrder loads a previously saved order that is not
// confirmed yet.
func (otx OrderTx) RetrieveOrder(orderID string) (paywall.Subscription, error) {
	var subs paywall.Subscription

	err := otx.tx.Get(
		&subs,
		otx.query.SelectSubsLock(),
		orderID,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.RetrieveOrder").Error(err)

		return subs, err
	}

	return subs, nil
}

// ConfirmOrder set an order's confirmation time.
func (otx OrderTx) ConfirmOrder(order paywall.Subscription) error {
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
func (otx OrderTx) FindBalanceSources(accountID paywall.AccountID) ([]paywall.BalanceSource, error) {
	rows, err := otx.tx.Query(
		otx.query.BalanceSource(),
		accountID.CompoundID,
		accountID.UnionID)

	if err != nil {
		logger.WithField("trace", "OrderTx.FindBalanceSources").Error(err)
		return nil, err
	}
	defer rows.Close()

	orders := make([]paywall.BalanceSource, 0)
	for rows.Next() {
		var o paywall.BalanceSource

		err := rows.Scan(
			&o.ID,
			&o.NetPrice,
			&o.StartDate,
			&o.EndDate)

		if err != nil {
			logger.WithField("trace", "OrderTx.FindBalanceSources").Error(err)
			return nil, err
		}

		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		logger.WithField("trace", "OrderTx.FindBalanceSources").Error(err)
		return nil, err
	}

	return orders, nil
}

// SaveUpgrade saves the data about upgrading.
// Deprecate
func (otx OrderTx) SaveUpgrade(orderID string, up paywall.Upgrade) error {
	_, err := otx.tx.Exec(otx.query.InsertUpgrade(),
		up.ID,
		orderID,
		up.Balance,
		up.SourceOrderIDs(),
		up.Member.ID,
		up.Member.Cycle,
		up.Member.ExpireDate,
		up.Member.User.FtcID,
		up.Member.User.UnionID,
		up.Member.Tier)

	if err != nil {
		return err
	}

	return nil
}

func (otx OrderTx) SaveUpgradeV2(up paywall.UpgradePreview, m paywall.Membership) error {

	var data = struct {
		paywall.UpgradePreview
		paywall.Membership
	}{
		UpgradePreview: up,
		Membership:     m,
	}
	_, err := otx.tx.NamedExec(
		otx.query.InsertUpgrade(),
		data)

	if err != nil {
		return err
	}

	return nil
}

// SetUpgradeTarget set the upgrade id on all rows that can
// be used as balance source.
// This operation should be performed together with
// SaveOrder and SaveUpgrade.
// Deprecate
func (otx OrderTx) SetLastUpgradeID(up paywall.Upgrade) error {
	strList := strings.Join(up.Source, ",")
	_, err := otx.tx.Exec(otx.query.SetLastUpgradeID(),
		up.ID,
		strList)

	if err != nil {
		return err
	}

	return nil
}

// SetLastUpgradeIDV2 set the last_upgrade_id on an order
// so that balance used for proration cannot be used next
// time.
func (otx OrderTx) SetLastUpgradeIDV2(up paywall.UpgradePreview) error {

	_, err := otx.tx.Exec(otx.query.SetLastUpgradeID(),
		up.ID,
		up.SourceOrderIDs())

	if err != nil {
		return err
	}

	return nil
}

// ConfirmUpgrade set an upgrade's confirmation time.
func (otx OrderTx) ConfirmUpgrade(id string) error {
	_, err := otx.tx.Exec(otx.query.ConfirmUpgrade(), id)
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
