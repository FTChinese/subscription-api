package model

import (
	"database/sql"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"strings"
)

// MemberTx confirm a payment and creates/renews/upgrades
// membership based on the payment result, all done in one
// transaction.
type MemberTx struct {
	tx    *sql.Tx
	query query.Builder
}

// RetrieveOrder loads a previously saved order.
func (m MemberTx) RetrieveOrder(orderID string) (paywall.Subscription, error) {
	var subs paywall.Subscription
	var ids string

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
		&ids,
		&subs.PaymentMethod,
		&subs.ConfirmedAt,
		&subs.IsConfirmed,
	)

	subs.ProrationSource = strings.Split(ids, ",")

	if err != nil {
		logger.WithField("trace", "MemberTx.RetrieveOrder").Error(err)

		_ = m.tx.Rollback()

		return subs, err
	}

	// Already confirmed.
	if subs.IsConfirmed {
		logger.WithField("trace", "MemberTx.RetrieveOrder").Infof("Order %s is already confirmed", orderID)

		_ = m.tx.Rollback()

		return subs, ErrAlreadyConfirmed
	}

	return subs, nil
}

func (m MemberTx) RetrieveMember(subs paywall.Subscription) (paywall.Membership, error) {
	var member paywall.Membership

	err := m.tx.QueryRow(
		m.query.SelectMemberLock(),
		subs.CompoundID,
		subs.UnionID,
	).Scan(
		&member.CompoundID,
		&member.UnionID,
		&member.Tier,
		&member.Cycle,
		&member.ExpireDate,
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

func (m MemberTx) InvalidUpgrade(orderID string, errInvalid error) {
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
		_ = m.tx.Rollback()
		return
	}

	_ = m.tx.Rollback()

	return
}

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

func (m MemberTx) MarkOrdersProrated(subs paywall.Subscription) error {

	orderIDs := strings.Join(subs.ProrationSource, ",")

	_, err := m.tx.Exec(
		m.query.Prorated(),
		subs.OrderID,
		subs.CompoundID,
		subs.UnionID,
		orderIDs)

	if err != nil {
		logger.WithField("trace", "MemberTx.MarkOrdersProrated").Error(err)
		_ = m.tx.Rollback()
		return err
	}

	return nil
}

func (m MemberTx) UpsertMember(mm paywall.Membership) error {
	_, err := m.tx.Exec(
		m.query.UpsertMember(),
		mm.CompoundID,
		mm.UnionID,
		mm.FTCUserID,
		mm.UnionID,
		mm.Tier,
		mm.Cycle,
		mm.ExpireDate,
		mm.FTCUserID,
		mm.UnionID,
		mm.Tier,
		mm.Cycle,
		mm.ExpireDate,
	)

	if err != nil {
		logger.WithField("trace", "MemberTx.UpsertMeber").Error(err)
		_ = m.tx.Rollback()

		return err
	}

	return nil
}

func (m MemberTx) Commit() error {
	return m.tx.Commit()
}
