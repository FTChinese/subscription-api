package model

import (
	"database/sql"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
)

type StripeTx struct {
	tx    *sql.Tx
	query query.Builder
}

// Retrieves a member. Empty Membership value is taken as valid.
func (t StripeTx) RetrieveMember(ftcID string) (paywall.Membership, error) {
	var m paywall.Membership

	err := t.tx.QueryRow(
		t.query.SelectMemberLock(),
		ftcID,
		nil,
	).Scan(
		&m.ID,
		&m.CompoundID,
		&m.UnionID,
		&m.Tier,
		&m.Cycle,
		&m.ExpireDate,
		&m.PaymentMethod,
		&m.StripeSubID,
		&m.AutoRenewal)

	if err != nil && err != sql.ErrNoRows {
		logger.WithField("trace", "StripeTx.RetrieveMember").Error(err)

		return m, err
	}

	return m, nil
}

func (t StripeTx) CreateMember(m paywall.Membership, planID string) error {
	vipType := tierID(m.Tier)
	expireTime := m.ExpireDate.Unix()

	_, err := t.tx.Exec(
		t.query.InsertStripeMember(),
		m.ID,
		m.CompoundID,
		m.UnionID,
		vipType,
		expireTime,
		m.FTCUserID,
		m.UnionID,
		m.Tier,
		m.Cycle,
		m.ExpireDate,
		m.PaymentMethod,
		m.StripeSubID,
		planID,
		m.AutoRenewal,
	)

	if err != nil {
		logger.WithField("trace", "StripeTx.CreateMember").Error(err)
		return err
	}

	return nil
}

func (t StripeTx) UpdateMember(m paywall.Membership, planID string) error {
	vipType := tierID(m.Tier)
	expireTime := m.ExpireDate.Unix()

	_, err := t.tx.Exec(t.query.UpdateStripeMember(),
		m.ID,
		vipType,
		expireTime,
		m.Tier,
		m.Cycle,
		m.ExpireDate,
		m.PaymentMethod,
		m.StripeSubID,
		planID,
		m.AutoRenewal)

	if err != nil {
		logger.WithField("trace", "StripeTx.UpdateMembership").Error(err)
		return err
	}

	return nil
}

// LinkUser adds membership id to user table.
func (t StripeTx) LinkUser(m paywall.Membership) error {
	if m.IsFtc() {
		_, err := t.tx.Exec(t.query.LinkFtcMember(),
			m.ID,
			m.FTCUserID)
		if err != nil {
			return err
		}
	}

	if m.IsWx() {
		_, err := t.tx.Exec(t.query.LinkWxMember(),
			m.ID,
			m.UnionID)

		if err != nil {
			return err
		}
	}

	return nil
}

func (t StripeTx) rollback() error {
	return t.tx.Rollback()
}

func (t StripeTx) commit() error {
	return t.tx.Commit()
}
