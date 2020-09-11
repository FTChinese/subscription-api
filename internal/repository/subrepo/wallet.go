package subrepo

import "github.com/FTChinese/subscription-api/pkg/subs"

func (env Env) ListProratedOrders(upOrderID string) ([]subs.ProratedOrder, error) {
	var po = make([]subs.ProratedOrder, 0)

	err := env.db.Select(
		&po,
		subs.StmtListProratedOrders,
		upOrderID)

	if err != nil {
		logger.WithField("trace", "SubEnv.RetrieveProratedOrders").Error(err)
		return nil, err
	}

	return po, nil
}
