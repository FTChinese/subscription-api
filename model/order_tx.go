package model

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
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
func (t OrderTx) RetrieveMember(u paywall.UserID) (paywall.Membership, error) {
	var m paywall.Membership

	// In the ftc_vip table, vip_id might be ftc uuid or wechat
	// id, and vip_id_alias is always wechat id.
	// In future, the table will be refactor with two extra
	// columns ftc_user_id dedicated to ftc uuid and wx_union_id
	// dedicated for wechat union id. The vip_id column will be
	// use only as a unique constraint on these two columns.
	err := t.tx.QueryRow(
		t.query.SelectMemberLock(),
		u.CompoundID,
		u.UnionID,
	).Scan(
		&m.ID,
		&m.CompoundID,
		&m.UnionID,
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

	plan, err := paywall.GetFtcPlans(t.live).FindPlan(m.PlanID())
	if err == nil {
		m.FtcPlan = plan
	}

	// Treat a non-existing member as a valid value.
	return m, nil
}

// SaveOrder saves an order to db.
func (t OrderTx) SaveOrder(s paywall.Subscription, c util.ClientApp) error {

	_, err := t.tx.Exec(
		t.query.InsertSubs(),
		s.ID,
		s.CompoundID,
		s.FtcID,
		s.UnionID,
		s.ListPrice,
		s.Amount,
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

// FindBalanceSources retrieves all orders that has unused portions.
// Used to build upgrade order for alipay and wxpay
func (t OrderTx) FindBalanceSources(userID paywall.UserID) ([]paywall.BalanceSource, error) {
	rows, err := t.tx.Query(
		t.query.BalanceSource(),
		userID.CompoundID,
		userID.UnionID)
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
func (t OrderTx) SaveUpgrade(orderID string, up paywall.Upgrade) error {
	_, err := t.tx.Exec(t.query.InsertUpgrade(),
		up.ID,
		orderID,
		up.Balance,
		up.SourceOrderIDs(),
		up.Member.ID,
		up.Member.Cycle,
		up.Member.ExpireDate,
		up.Member.FtcID,
		up.Member.UnionID,
		up.Member.Tier)

	if err != nil {
		return err
	}

	return nil
}

func (t OrderTx) SaveUpgradeV2(orderID string, up paywall.UpgradePreview) error {
	_, err := t.tx.Exec(t.query.InsertUpgrade(),
		up.ID,
		orderID,
		up.Balance,
		up.SourceOrderIDs(),
		up.Member.ID,
		up.Member.Cycle,
		up.Member.ExpireDate,
		up.Member.FtcID,
		up.Member.UnionID,
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
func (t OrderTx) SetUpgradeIDOnSource(up paywall.Upgrade) error {
	strList := strings.Join(up.Source, ",")
	_, err := t.tx.Exec(t.query.SetUpgradeIDOnSource(),
		up.ID,
		strList)

	if err != nil {
		return err
	}

	return nil
}

func (t OrderTx) SetUpgradeIDOnSourceV2(up paywall.UpgradePreview) error {

	_, err := t.tx.Exec(t.query.SetUpgradeIDOnSource(),
		up.ID,
		up.SourceOrderIDs())

	if err != nil {
		return err
	}

	return nil
}

// ----------------------------------
// The following only applicable to confirmation of orders
func tierID(tier enum.Tier) int64 {
	switch tier {
	case enum.TierStandard:
		return 10
	case enum.TierPremium:
		return 100
	}

	return 0
}

// RetrieveOrder loads a previously saved order that is not
// confirmed yet.
func (t OrderTx) RetrieveOrder(orderID string) (paywall.Subscription, error) {
	var subs paywall.Subscription

	err := t.tx.QueryRow(
		t.query.SelectSubsLock(),
		orderID,
	).Scan(
		&subs.ID,
		&subs.CompoundID,
		&subs.FtcID,
		&subs.UnionID,
		&subs.ListPrice,
		&subs.Amount,
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

	// Already confirmed.
	if subs.IsConfirmed {
		logger.WithField("trace", "MemberTx.RetrieveOrder").Infof("Order %s is already confirmed", orderID)

		return subs, util.ErrAlreadyConfirmed
	}

	return subs, nil
}

// ConfirmOrder set an order's confirmation time.
func (t OrderTx) ConfirmOrder(order paywall.Subscription) error {
	_, err := t.tx.Exec(
		t.query.ConfirmSubs(),
		order.ConfirmedAt,
		order.StartDate,
		order.EndDate,
		order.ID,
	)

	if err != nil {
		logger.WithField("trace", "OrderTx.ConfirmOrder").Error(err)
		_ = t.tx.Rollback()
		return err
	}

	return nil
}

// ConfirmUpgrade set an upgrade's confirmation time.
func (t OrderTx) ConfirmUpgrade(id string) error {
	_, err := t.tx.Exec(t.query.ConfirmUpgrade(), id)
	if err != nil {
		logger.WithField("trace", "OrderTx.ConfirmUpgrade").Error(err)
		return err
	}

	return nil
}

func (t OrderTx) DuplicateUpgrade(orderID string) error {
	_, err := t.tx.Exec(
		t.query.UpgradeFailure(),
		"failed",
		"duplicate_upgrade")

	if err != nil {
		logger.WithField("trace", "MemberTx.InvalidUpgrade").Error(err)
		return err
	}

	return nil
}

func (t OrderTx) CreateMember(m paywall.Membership) error {
	vipType := tierID(m.Tier)
	expireTime := m.ExpireDate.Unix()

	_, err := t.tx.Exec(
		t.query.InsertMember(),
		m.ID,
		m.CompoundID,
		m.UnionID,
		vipType,
		expireTime,
		m.FtcID,
		m.UnionID,
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

func (t OrderTx) UpdateMember(m paywall.Membership) error {
	vipType := tierID(m.Tier)
	expireTime := m.ExpireDate.Unix()

	_, err := t.tx.Exec(t.query.UpdateMember(),
		m.ID,
		vipType,
		expireTime,
		m.Tier,
		m.Cycle,
		m.ExpireDate,
		m.PaymentMethod,
		m.StripeSubID,
		m.StripePlanID,
		m.AutoRenewal,
		m.CompoundID,
		m.Status,
		m.UnionID)

	if err != nil {
		logger.WithField("trace", "OrderTx.UpdateMembership").Error(err)
		return err
	}

	return nil
}

// -------------
// The following are used by gift card

func (t OrderTx) ActivateGiftCard(code string) error {
	_, err := t.tx.Exec(t.query.ActivateGiftCard(), code)

	if err != nil {
		return err
	}

	return nil
}

func (t OrderTx) rollback() error {
	return t.tx.Rollback()
}

func (t OrderTx) commit() error {
	return t.tx.Commit()
}
