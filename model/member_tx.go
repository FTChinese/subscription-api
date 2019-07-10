package model

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
)

// MemberTx confirm a payment and creates/renews/upgrades
// membership based on the payment result, all done in one
// transaction.
type MemberTx struct {
	tx    *sql.Tx
	query query.Builder
}

// RetrieveOrder loads a previously saved order that is not
// confirmed yet.
func (m MemberTx) RetrieveOrder(orderID string) (paywall.Subscription, error) {
	var subs paywall.Subscription

	err := m.tx.QueryRow(
		m.query.SelectSubsLock(),
		orderID,
	).Scan(
		&subs.OrderID,
		&subs.CompoundID,
		&subs.FtcID,
		&subs.UnionID,
		&subs.ListPrice,
		&subs.NetPrice,
		&subs.TierToBuy,
		&subs.BillingCycle,
		&subs.CycleCount,
		&subs.ExtraDays,
		&subs.Kind,
		&subs.PaymentMethod,
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

		return subs, ErrAlreadyConfirmed
	}

	return subs, nil
}

// RetrieveUpgradeSource loads the order ids upon which an
// upgrade order is build.
func (m MemberTx) RetrieveUpgradeSource(upgradeID string) ([]string, error) {
	rows, err := m.tx.Query(m.query.SelectUpgradeSource(),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "MemberTx.RetrieveUpgradeSource").Error(err)
		return nil, err
	}

	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string

		err := rows.Scan(&id)

		if err != nil {
			logger.WithField("trace", "MemberTx.RetrieveUpgradeSource").Error(err)
			return nil, err
		}

		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		logger.WithField("trace", "MemberTx.RetrieveUpgradeSource").Error(err)
		return nil, err
	}

	return ids, nil
}

// RetrieveMember find whether an order is created by an
// existing member.
func (m MemberTx) RetrieveMember(id paywall.UserID) (paywall.Membership, error) {
	var member paywall.Membership

	err := m.tx.QueryRow(
		m.query.SelectMemberLock(),
		id.CompoundID,
		id.UnionID,
	).Scan(
		&member.ID,
		&member.CompoundID,
		&member.UnionID,
		&member.Tier,
		&member.Cycle,
		&member.ExpireDate,
		&member.PaymentMethod,
		&member.StripeSubID,
		&member.AutoRenewal,
	)

	if err != nil && err != sql.ErrNoRows {
		logger.WithField("trace", "MemberTx.RetrieveMember").Error(err)

		_ = m.tx.Rollback()
		return member, err
	}

	// The value return by ErrNoRows is valid.
	// A empty member's Tier is enum.InvalidTier.
	// Caller can use this value to determine if member exists
	// or not.
	return member, nil
}

// InvalidUpgrade marks an upgrade order as invalid.
// Transaction should rollback regardless of success or not
// since this action itself is used to handle error.
func (m MemberTx) InvalidUpgrade(orderID string, errInvalid error) error {
	var reason string
	switch errInvalid {
	case paywall.ErrDuplicateUpgrading:
		reason = "duplicate"
	case paywall.ErrNoUpgradingTarget:
		reason = "missing_target"
	}

	_, err := m.tx.Exec(
		m.query.InvalidUpgrade(),
		reason,
		orderID)

	if err != nil {
		logger.WithField("trace", "MemberTx.InvalidUpgrade").Error(err)
		return err
	}

	return nil
}

// ConfirmOrder set an order's confirmation time
func (m MemberTx) ConfirmOrder(subs paywall.Subscription) error {
	_, err := m.tx.Exec(
		m.query.ConfirmSubs(),
		subs.ConfirmedAt,
		subs.StartDate,
		subs.EndDate,
		subs.OrderID,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.ConfirmOrder").Error(err)
		_ = m.tx.Rollback()
		return err
	}

	return nil
}

// ConfirmUpgradeSource set all orders used for upgrading
// as confirmed, using the upgrading order's id.
func (m MemberTx) ConfirmUpgradeSource(upID string) error {
	_, err := m.tx.Exec(m.query.ConfirmUpgradeSource(),
		upID)

	if err != nil {
		logger.WithField("trace", "MemberTx.ConfirmUpgradeSource").Error(err)
		return err
	}

	return err
}

func tierID(tier enum.Tier) int64 {
	switch tier {
	case enum.TierStandard:
		return 10
	case enum.TierPremium:
		return 100
	}

	return 0
}

// UpsertMember creates or update a member.
func (m MemberTx) UpsertMember(mm paywall.Membership) error {

	vipType := tierID(mm.Tier)
	expireTime := mm.ExpireDate.Unix()

	_, err := m.tx.Exec(
		m.query.UpsertMember(),
		mm.CompoundID,
		mm.UnionID,
		vipType,
		expireTime,
		mm.FTCUserID,
		mm.UnionID,
		mm.Tier,
		mm.Cycle,
		mm.ExpireDate,
		mm.PaymentMethod,
		mm.StripeSubID,
		mm.AutoRenewal,
		mm.CompoundID,
		mm.UnionID,
		vipType,
		expireTime,
		mm.FTCUserID,
		mm.UnionID,
		mm.Tier,
		mm.Cycle,
		mm.ExpireDate,
		mm.PaymentMethod,
		mm.StripeSubID,
		mm.AutoRenewal,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.UpsertMember").Error(err)
		_ = m.tx.Rollback()

		return err
	}

	return nil
}

// LinkUser adds membership id to user table.
func (m MemberTx) LinkUser(mm paywall.Membership) error {
	if mm.IsFtc() {
		_, err := m.tx.Exec(m.query.LinkFtcMember(),
			mm.ID,
			mm.FTCUserID)
		if err != nil {
			return err
		}
	}

	if mm.IsWx() {
		_, err := m.tx.Exec(m.query.LinkWxMember(),
			mm.ID,
			mm.UnionID)

		if err != nil {
			return err
		}
	}

	return nil
}

func (m MemberTx) rollback() error {
	return m.tx.Rollback()
}

func (m MemberTx) commit() error {
	return m.tx.Commit()
}
