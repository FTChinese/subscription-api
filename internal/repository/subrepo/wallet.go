package subrepo

import "github.com/FTChinese/subscription-api/pkg/subs"

func (env Env) ListProratedOrders(upOrderID string) ([]subs.ProratedOrder, error) {
	var po = make([]subs.ProratedOrder, 0)

	err := env.rwdDB.Select(
		&po,
		subs.StmtListProratedOrders,
		upOrderID)

	if err != nil {
		return nil, err
	}

	return po, nil
}
