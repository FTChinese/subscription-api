package model

import (
	"database/sql"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/util"
	"strings"
)

// OrderTx check a user's member status and create an order
// if allowed.
type OrderTx struct {
	tx    *sql.Tx
	live  bool
	query query.Builder
}

// RetrieveMember retrieves a user's membership info by ftc id
// or wechat union id.
// Returns zero value of membership if not found.
func (otx OrderTx) RetrieveMember(id paywall.AccountID) (paywall.Membership, error) {
	var m paywall.Membership

	// In the ftc_vip table, vip_id might be ftc uuid or wechat
	// id, and vip_id_alias is always wechat id.
	// In future, the table will be refactor with two extra
	// columns ftc_user_id dedicated to ftc uuid and wx_union_id
	// dedicated for wechat union id. The vip_id column will be
	// use only as a unique constraint on these two columns.
	err := otx.tx.QueryRow(
		otx.query.SelectMemberLock(),
		id.CompoundID,
		id.UnionID,
	).Scan(
		&m.ID,
		&m.User.CompoundID,
		&m.User.FtcID,
		&m.User.UnionID,
		&m.Tier,
		&m.Cycle,
		&m.ExpireDate,
		&m.PaymentMethod,
		&m.StripeSubID,
		&m.AutoRenewal,
		&m.Status)

	if err != nil && err != sql.ErrNoRows {
		logger.WithField("trace", "OrderTx.RetrieveMember").Error(err)

		return m, err
	}

	// Treat a non-existing member as a valid value.
	return m, nil
}

// SaveOrder saves an order to db.
// This is only limited to alipay and wechat pay.
// Stripe pay does not generate any orders on our side.
func (otx OrderTx) SaveOrder(s paywall.Subscription, c util.ClientApp) error {

	_, err := otx.tx.Exec(
		otx.query.InsertSubs(),
		s.ID,
		s.User.CompoundID,
		s.User.FtcID,
		s.User.UnionID,
		s.ListPrice,
		s.NetPrice,
		s.Tier,
		s.Cycle,
		s.CycleCount,
		s.ExtraDays,
		s.Usage,
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

// RetrieveOrder loads a previously saved order that is not
// confirmed yet.
func (otx OrderTx) RetrieveOrder(orderID string) (paywall.Subscription, error) {
	var subs paywall.Subscription

	err := otx.tx.QueryRow(
		otx.query.SelectSubsLock(),
		orderID,
	).Scan(
		&subs.ID,
		&subs.User.CompoundID,
		&subs.User.FtcID,
		&subs.User.UnionID,
		&subs.ListPrice,
		&subs.NetPrice,
		&subs.Tier,
		&subs.Cycle,
		&subs.CycleCount,
		&subs.ExtraDays,
		&subs.Usage,
		&subs.PaymentMethod,
		&subs.CreatedAt,
		&subs.ConfirmedAt,
		&subs.IsConfirmed,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.RetrieveOrder").Error(err)

		return subs, err
	}

	return subs, nil
}

// ConfirmOrder set an order's confirmation time.
func (otx OrderTx) ConfirmOrder(order paywall.Subscription) error {
	_, err := otx.tx.Exec(
		otx.query.ConfirmSubs(),
		order.ConfirmedAt,
		order.StartDate,
		order.EndDate,
		order.ID,
	)

	if err != nil {
		logger.WithField("trace", "OrderTx.ConfirmOrder").Error(err)

		return err
	}

	return nil
}

func (otx OrderTx) CreateMember(m paywall.Membership) error {

	_, err := otx.tx.Exec(
		otx.query.InsertMember(),
		m.ID,
		m.User.CompoundID,
		m.User.UnionID,
		m.TierCode(),
		m.ExpireDate.Unix(),
		m.User.FtcID,
		m.User.UnionID,
		m.Tier,
		m.Cycle,
		m.ExpireDate,
		m.PaymentMethod,
		m.StripeSubID,
		m.StripePlanID,
		m.AutoRenewal,
		m.Status,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.CreateMember").Error(err)
		return err
	}

	return nil
}

func (otx OrderTx) UpdateMember(m paywall.Membership) error {

	_, err := otx.tx.Exec(otx.query.UpdateMember(),
		m.ID,
		m.TierCode(),
		m.ExpireDate.Unix(),
		m.Tier,
		m.Cycle,
		m.ExpireDate,
		m.PaymentMethod,
		m.StripeSubID,
		m.StripePlanID,
		m.AutoRenewal,
		m.Status,
		m.User.CompoundID,
		m.User.UnionID)

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

func (otx OrderTx) SaveUpgradeV2(orderID string, up paywall.UpgradePreview) error {
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

// SetUpgradeTarget set the upgrade id on all rows that can
// be used as balance source.
// This operation should be performed together with
// SaveOrder and SaveUpgrade.
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

func (otx OrderTx) rollback() error {
	return otx.tx.Rollback()
}

func (otx OrderTx) commit() error {
	return otx.tx.Commit()
}
