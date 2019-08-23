package repository

import "gitlab.com/ftchinese/subscription-api/models/paywall"

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

func (env Env) BackUpMember(m paywall.MemberSnapshot) error {
	_, err := env.db.NamedExec(
		env.query.MemberSnapshot(),
		m)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) RetrieveMember(id paywall.AccountID) (paywall.Membership, error) {
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

func (env Env) FindUnusedOrders(id paywall.AccountID) ([]paywall.ProrationSource, error) {
	var sources = []paywall.ProrationSource{}

	err := env.db.Select(
		&sources,
		env.query.SelectProrationSource(),
		id.CompoundID,
		id.UnionID)

	if err != nil {
		logger.WithField("trace", "Env.FindUnusedOrders").Error(err)
		return sources, err
	}

	return sources, nil
}

func (env Env) RetrieveUpgradePlan(upgradeID string) (paywall.UpgradePlan, error) {
	var up paywall.UpgradePlan

	err := env.db.Get(
		&up,
		env.query.SelectUpgradePlan(),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "Env.RetrieveUpgradePlan").Error(err)
		return up, err
	}

	return up, nil
}

func (env Env) RetrieveUpgradeSource(upgradeID string) ([]paywall.ProrationSource, error) {
	var sources = []paywall.ProrationSource{}

	err := env.db.Select(
		&sources,
		env.query.SelectUpgradeSource(),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "Env.RetrieveUpgradeSource").Error(err)
		return sources, err
	}

	return sources, nil
}
