package readerrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

// RetrieveUpgradeBalance retrieves an upgrade plan to be used in email sent to user.
func (env ReaderEnv) retrieveUpgradeWallet(upgradeID string) (subscription.UpgradeSchema, error) {

	var data subscription.UpgradeSchema

	err := env.db.Get(
		&data,
		query.BuildSelectUpgradePlan(env.UseSandboxDB()),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "SubEnv.RetrieveUpgradeWallet").Error(err)
		return subscription.UpgradeSchema{}, err
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
	upgradeSchema, err := env.retrieveUpgradeWallet(upgradeID)
	if err != nil {
		return subscription.UpgradeSchema{}, err
	}

	sources, err := env.retrieveProratedOrders(upgradeID)
	if err != nil {
		return subscription.UpgradeSchema{}, err
	}

	upgradeSchema.Sources = sources
	return upgradeSchema, nil
}
