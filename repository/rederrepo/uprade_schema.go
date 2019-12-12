package rederrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

// RetrieveUpgradeBalance retrieves an upgrade plan to be used in email sent to user.
func (env ReaderEnv) retrieveUpgradeBalance(upgradeID string) (subscription.UpgradeBalanceSchema, error) {

	var data subscription.UpgradeBalanceSchema

	err := env.db.Get(
		&data,
		query.BuildSelectUpgradePlan(env.UseSandboxDB()),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "SubEnv.RetrieveUpgradeBalance").Error(err)
		return subscription.UpgradeBalanceSchema{}, err
	}

	return data, nil
}

// RetrieveProratedOrders retrieves all orders prorated from
// proration table. Used to send user an email after upgrade.
func (env ReaderEnv) retrieveProratedOrders(upgradeID string) ([]subscription.ProratedOrderSchema, error) {
	var sources = make([]subscription.ProratedOrderSchema, 0)

	err := env.db.Select(
		&sources,
		query.BuildSelectProration(env.UseSandboxDB()),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "SubEnv.RetrieveProratedOrders").Error(err)
		return sources, err
	}

	return sources, nil
}

func (env ReaderEnv) LoadUpgradeSchema(upgradeID string) (subscription.UpgradeSchema, error) {
	balance, err := env.retrieveUpgradeBalance(upgradeID)
	if err != nil {
		return subscription.UpgradeSchema{}, err
	}

	sources, err := env.retrieveProratedOrders(upgradeID)
	if err != nil {
		return subscription.UpgradeSchema{}, err
	}

	return subscription.UpgradeSchema{
		UpgradeBalanceSchema: balance,
		Sources:              sources,
	}, nil
}
