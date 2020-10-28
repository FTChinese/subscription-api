package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

func (env Env) CreateOrder(config subs.PaymentConfig) (subs.Order, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	otx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return subs.Order{}, err
	}

	// Step 1: Retrieve membership for this user.
	// The membership might be empty but the value is
	// valid.
	sugar.Infof("Start retrieving membership for reader %+v", config.Account.MemberID())
	member, err := otx.RetrieveMember(config.Account.MemberID())
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.Order{}, err
	}
	sugar.Infof("Membership retrieved %+v", member)

	// Deduce order kind.
	kind, err := member.AliWxSubsKind(config.Plan.Edition)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.Order{}, err
	}
	sugar.Infof("Subscription kind %s", kind)

	// Step 2: Build an order for the user's chosen plan
	// with chosen payment method based on previous
	// membership so that we could how this order
	// is used: create, renew or upgrade.

	var balanceSources []subs.BalanceSource
	// Step 3: required only if this order is used for
	// upgrading.
	if kind == enum.OrderKindUpgrade {
		// Step 3.1: find previous orders with balance
		// remaining.
		// DO not save sources directly. The balance is not
		// calculated at this point.
		sugar.Infof("Get balance sources for an upgrading order")
		balanceSources, err = otx.FindBalanceSources(config.Account.MemberID())
		if err != nil {
			sugar.Error(err)
			_ = otx.Rollback()
			return subs.Order{}, err
		}
		sugar.Infof("Find balance source: %+v", balanceSources)
	}

	checkout := config.Checkout(balanceSources, kind)

	order, err := config.BuildOrder(checkout)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.Order{}, err
	}

	// Step 4: Save this order.
	if err := otx.SaveOrder(order); err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.Order{}, err
	}
	sugar.Infof("Order saved %s", order.ID)

	// Step 5: Save prorated orders for upgrade.
	if kind == enum.OrderKindUpgrade {
		err := otx.SaveProratedOrders(checkout.ProratedOrders(order))
		if err != nil {
			sugar.Error(err)
			_ = otx.Rollback()
			return subs.Order{}, err
		}
	}

	if err := otx.Commit(); err != nil {
		sugar.Error(err)
		return subs.Order{}, err
	}

	return order, nil
}

func (env Env) RetrieveOrder(orderID string) (subs.Order, error) {
	var order subs.Order

	err := env.db.Get(
		&order,
		subs.StmtSelectOrder,
		orderID)

	if err != nil {
		return subs.Order{}, err
	}

	return order, nil
}
