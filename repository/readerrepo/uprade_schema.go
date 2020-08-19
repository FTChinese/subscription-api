package readerrepo

import (
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// retrieveUpgradeBalance retrieves an upgrade schema for an order.
func (env ReaderEnv) retrieveUpgradeBalance(orderID string) (subs.UpgradeSchema, error) {

	var s subs.UpgradeSchema

	err := env.db.Get(
		&s,
		subs.StmtUpgradeBalance(env.GetSubsDB()),
		orderID)

	if err != nil {
		logger.WithField("trace", "SubEnv.RetrieveUpgradeBalance").Error(err)
		return subs.UpgradeSchema{}, err
	}

	return s, nil
}

// retrieveProratedOrders retrieves all orders prorated used in an upgrade session.
func (env ReaderEnv) retrieveProratedOrders(upgradeID string) ([]subs.ProratedOrderSchema, error) {
	var sources = make([]subs.ProratedOrderSchema, 0)

	err := env.db.Select(
		&sources,
		subs.StmtListProratedOrders(env.GetSubsDB()),
		upgradeID)

	if err != nil {
		logger.WithField("trace", "SubEnv.RetrieveProratedOrders").Error(err)
		return sources, err
	}

	return sources, nil
}

func (env ReaderEnv) LoadUpgradeSchema(orderID string) (subs.UpgradeSchema, error) {
	upgradeSchema, err := env.retrieveUpgradeBalance(orderID)
	if err != nil {
		return subs.UpgradeSchema{}, err
	}

	sources, err := env.retrieveProratedOrders(upgradeSchema.ID)
	if err != nil {
		return subs.UpgradeSchema{}, err
	}

	upgradeSchema.Sources = sources
	return upgradeSchema, nil
}
