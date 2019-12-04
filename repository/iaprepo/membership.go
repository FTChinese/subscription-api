package iaprepo

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"gitlab.com/ftchinese/subscription-api/models/apple"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

type MembershipTx struct {
	tx      *sqlx.Tx
	sandbox bool
}

// RetrieveMember selects membership by compound id.
func (mtx MembershipTx) RetrieveMember(id reader.MemberID) (subscription.Membership, error) {
	var m subscription.Membership

	err := mtx.tx.Get(
		&m,
		query.BuildSelectMembership(mtx.sandbox, true),
		id.CompoundID)

	if err != nil && err != sql.ErrNoRows {
		logger.WithField("trace", "MembershipTx.RetrieveMember").Error(err)

		return m, err
	}

	m.Normalize()

	return m, nil
}

// RetrieveAppleMember selects membership by apple original transaction id.
func (mtx MembershipTx) RetrieveAppleMember(transactionID string) (subscription.Membership, error) {
	var m subscription.Membership

	err := mtx.tx.Get(
		&m,
		query.BuildSelectAppleMembership(mtx.sandbox),
		transactionID)

	if err != nil && err != sql.ErrNoRows {
		logger.WithField("trace", "MembershipTx.RetrieveMember").Error(err)

		return m, err
	}

	m.Normalize()

	return m, nil
}

// CreateMember inserts a membership row.
func (mtx MembershipTx) CreateMember(m subscription.Membership) error {
	m.Normalize()

	_, err := mtx.tx.NamedExec(
		query.BuildInsertMembership(mtx.sandbox),
		m)

	if err != nil {
		logger.WithField("trace", "MembershipTx.CreateMember").Error(err)
		return err
	}

	return nil
}

// UpdateMember updates an existing membership.
func (mtx MembershipTx) UpdateMember(m subscription.Membership) error {
	m.Normalize()

	_, err := mtx.tx.NamedExec(
		query.BuildUpdateMembership(mtx.sandbox),
		m)

	if err != nil {
		logger.WithField("trace", "MembershipTx.UpdateMember").Error(err)

		return err
	}

	return nil
}

func (mtx MembershipTx) DeleteMember(id reader.MemberID) error {
	_, err := mtx.tx.NamedExec(
		query.BuildDeleteMembership(mtx.sandbox),
		id)

	if err != nil {
		logger.WithField("trace", "MembershipTx.DeleteMember").Error(err)

		return err
	}

	return nil
}

func (mtx MembershipTx) Rollback() error {
	return mtx.tx.Rollback()
}

func (mtx MembershipTx) Commit() error {
	return mtx.tx.Commit()
}

// BackUpMember takes a snapshot of membership.
// TODO: handle apple's Environment.
func (env IAPEnv) BackUpMember(snapshot subscription.MemberSnapshot, e apple.Environment) error {
	_, err := env.db.NamedExec(
		query.BuildInsertMemberSnapshot(e == apple.EnvSandbox),
		snapshot)

	if err != nil {
		logger.WithField("trace", "IAPEnv.BackUpMember").Error(err)

		return err
	}

	return nil
}
