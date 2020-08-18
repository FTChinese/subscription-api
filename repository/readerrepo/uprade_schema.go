package readerrepo

import (
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// retrieveUpgradeBalance retrieves an upgrade plan to be used in email sent to user.
func (env ReaderEnv) retrieveUpgradeWallet(upgradeID string) (subs.UpgradeSchema, error) {

	var data subs.UpgradeSchema

	err := env.db.Get(
		&data,
		subs.StmtUpgradeBalance(env.GetSubsDB()),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "SubEnv.RetrieveUpgradeWallet").Error(err)
		return subs.UpgradeSchema{}, err
	}

	return data, nil
}

// RetrieveProratedOrders retrieves all orders prorated from
// proration table. Used to send user an email after upgrade.
func (env ReaderEnv) retrieveProratedOrders(upgradeID string) ([]subs.ProratedOrderSchema, error) {
	var sources = make([]subs.ProratedOrderSchema, 0)

	err := env.db.Select(
		&sources,
		subs.StmtListProration(env.GetSubsDB()),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "SubEnv.RetrieveProratedOrders").Error(err)
		return sources, err
	}

	return sources, nil
}

func (env ReaderEnv) LoadUpgradeSchema(upgradeID string) (subs.UpgradeSchema, error) {
	upgradeSchema, err := env.retrieveUpgradeWallet(upgradeID)
	if err != nil {
		return subs.UpgradeSchema{}, err
	}

	sources, err := env.retrieveProratedOrders(upgradeID)
	if err != nil {
		return subs.UpgradeSchema{}, err
	}

	upgradeSchema.Sources = sources
	return upgradeSchema, nil
}
