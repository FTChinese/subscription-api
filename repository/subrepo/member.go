package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

// AddMemberID set a membership's id column if it is empty.
func (env SubEnv) AddMemberID(m subscription.Membership) error {
	_, err := env.db.NamedExec(
		query.BuildUpdateMembershipID(env.UseSandboxDB()),
		m)

	if err != nil {
		return err
	}

	return nil
}

// BackUpMember saves a member's snapshot at a specific moment.
func (env SubEnv) BackUpMember(snapshot subscription.MemberSnapshot) error {
	_, err := env.db.NamedExec(
		query.BuildInsertMemberSnapshot(env.UseSandboxDB()),
		snapshot)

	if err != nil {
		return err
	}

	return nil
}

// FindBalanceSources creates a snapshot for orders with
// unused portion.
// This is identical to OrderTx.FindBalanceSources without a transaction.
func (env SubEnv) FindBalanceSources(id reader.MemberID) ([]subscription.ProratedOrderSchema, error) {
	var sources = make([]subscription.ProratedOrderSchema, 0)

	err := env.db.Select(
		&sources,
		query.BuildSelectBalanceSource(env.UseSandboxDB()),
		id.CompoundID,
		id.UnionID)

	if err != nil {
		logger.WithField("trace", "SubEnv.FindBalanceSources").Error(err)
		return sources, err
	}

	return sources, nil
}

// RetrieveUpgradeBalance retrieves an upgrade plan to be used in email sent to user.
//func (env SubEnv) RetrieveUpgradeBalance(upgradeID string) (subscription.UpgradeBalanceSchema, error) {
//
//	var data subscription.UpgradeBalanceSchema
//
//	err := env.db.Get(
//		&data,
//		query.BuildSelectUpgradePlan(env.UseSandboxDB()),
//		upgradeID)
//
//	if err != nil {
//		logger.WithField("trace", "SubEnv.RetrieveUpgradeBalance").Error(err)
//		return subscription.UpgradeBalanceSchema{}, err
//	}
//
//	return data, nil
//}

// RetrieveProratedOrders retrieves all orders prorated from
// proration table. Used to send user an email after upgrade.
//func (env SubEnv) RetrieveProratedOrders(upgradeID string) ([]subscription.ProratedOrderSchema, error) {
//	var sources = make([]subscription.ProratedOrderSchema, 0)
//
//	err := env.db.Select(
//		&sources,
//		query.BuildSelectProration(env.UseSandboxDB()),
//		upgradeID)
//
//	if err != nil {
//		logger.WithField("trace", "SubEnv.RetrieveProratedOrders").Error(err)
//		return sources, err
//	}
//
//	return sources, nil
//}
