package repository

import (
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/reader"
)

// AddMemberID set a membership's id column if it is empty.
func (env Env) AddMemberID(m paywall.Membership) error {
	_, err := env.db.NamedExec(
		env.query.AddMemberID(m.MemberColumn()),
		m)

	if err != nil {
		return err
	}

	return nil
}

// BackUpMember saves a member's snapshot at a specific moment.
func (env Env) BackUpMember(m paywall.MemberSnapshot) error {
	_, err := env.db.NamedExec(
		env.query.MemberSnapshot(),
		m)

	if err != nil {
		return err
	}

	return nil
}

// RetrieveMember retrieves membership from database.
func (env Env) RetrieveMember(id reader.AccountID) (paywall.Membership, error) {
	var m paywall.Membership

	err := env.db.Get(
		&m,
		env.query.SelectMember(id.MemberColumn()),
		id.CompoundID)

	if err != nil {
		return m, err
	}

	m.Normalize()

	return m, nil
}

// FindBalanceSources creates a snapshot for orders with
// unused portion.
// This is identical to OrderTx.FindBalanceSources without a transaction.
func (env Env) FindBalanceSources(id reader.AccountID) ([]paywall.ProrationSource, error) {
	var sources = []paywall.ProrationSource{}

	err := env.db.Select(
		&sources,
		env.query.SelectProrationSource(),
		id.CompoundID,
		id.UnionID)

	if err != nil {
		logger.WithField("trace", "Env.FindBalanceSources").Error(err)
		return sources, err
	}

	return sources, nil
}

// RetrieveUpgradePlan retrieves an upgrade plan to be used in email sent to user.
func (env Env) RetrieveUpgradePlan(upgradeID string) (paywall.UpgradePlan, error) {

	var data = struct {
		paywall.UpgradePlan
		paywall.Plan
	}{}

	err := env.db.Get(
		&data,
		env.query.SelectUpgradePlan(),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "Env.RetrieveUpgradePlan").Error(err)
		return paywall.UpgradePlan{}, err
	}

	return paywall.UpgradePlan{
		ID:        data.ID,
		Balance:   data.Balance,
		CreatedAt: data.CreatedAt,
		Data:      nil,
		Plan:      data.Plan,
	}, nil
}

// RetrieveProratedOrders retrieves all orders prorated from
// proration table. Used to send user an email after upgrade.
func (env Env) RetrieveProratedOrders(upgradeID string) ([]paywall.ProrationSource, error) {
	var sources = []paywall.ProrationSource{}

	err := env.db.Select(
		&sources,
		env.query.SelectProratedOrders(),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "Env.RetrieveProratedOrders").Error(err)
		return sources, err
	}

	return sources, nil
}
