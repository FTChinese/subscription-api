package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

// AddMemberID set a membership's id column if it is empty.
func (env SubEnv) AddMemberID(m subscription.Membership) error {
	_, err := env.db.NamedExec(
		query.BuildUpdateMembershipID(env.Sandbox),
		m)

	if err != nil {
		return err
	}

	return nil
}

// BackUpMember saves a member's snapshot at a specific moment.
func (env SubEnv) BackUpMember(snapshot subscription.MemberSnapshot) error {
	_, err := env.db.NamedExec(
		query.BuildInsertMemberSnapshot(env.Sandbox),
		snapshot)

	if err != nil {
		return err
	}

	return nil
}

// RetrieveMember retrieves membership from database.
func (env SubEnv) RetrieveMember(id reader.MemberID) (subscription.Membership, error) {
	var m subscription.Membership

	err := env.db.Get(
		&m,
		query.BuildSelectMembership(env.Sandbox, false),
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
func (env SubEnv) FindBalanceSources(id reader.MemberID) ([]subscription.ProrationSource, error) {
	var sources = []subscription.ProrationSource{}

	err := env.db.Select(
		&sources,
		query.BuildSelectBalanceSource(env.Sandbox),
		id.CompoundID,
		id.UnionID)

	if err != nil {
		logger.WithField("trace", "SubEnv.FindBalanceSources").Error(err)
		return sources, err
	}

	return sources, nil
}

// RetrieveUpgradePlan retrieves an upgrade plan to be used in email sent to user.
func (env SubEnv) RetrieveUpgradePlan(upgradeID string) (subscription.UpgradeIntent, error) {

	var data = struct {
		subscription.UpgradeIntent
		plan.Plan
	}{}

	err := env.db.Get(
		&data,
		query.BuildSelectUpgradePlan(env.Sandbox),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "SubEnv.RetrieveUpgradePlan").Error(err)
		return subscription.UpgradeIntent{}, err
	}

	return subscription.UpgradeIntent{
		ID:        data.ID,
		Balance:   data.Balance,
		CreatedAt: data.CreatedAt,
		Data:      nil,
		Plan:      data.Plan,
	}, nil
}

// RetrieveProratedOrders retrieves all orders prorated from
// proration table. Used to send user an email after upgrade.
func (env SubEnv) RetrieveProratedOrders(upgradeID string) ([]subscription.ProrationSource, error) {
	var sources = []subscription.ProrationSource{}

	err := env.db.Select(
		&sources,
		query.BuildSelectProration(env.Sandbox),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "SubEnv.RetrieveProratedOrders").Error(err)
		return sources, err
	}

	return sources, nil
}
