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
		&subs.UpgradeID,
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

func (m MemberTx) DuplicateUpgrade(orderID string) error {
	_, err := m.tx.Exec(
		m.query.UpgradeFailure(),
		"failed",
		"duplicate_upgrade")

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

func (m MemberTx) ConfirmUpgrade(id string) error {
	_, err := m.tx.Exec(m.query.ConfirmUpgrade(), id)
	if err != nil {
		logger.WithField("trace", "MemberTx.ConfirmUpgrade").Error(err)
		return err
	}

	return nil
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
