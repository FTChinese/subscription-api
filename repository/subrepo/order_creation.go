package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"go.uber.org/zap"
	"time"
)

func (env Env) CreateOrder(builder *subs.OrderBuilder) (subs.Order, error) {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	otx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return subs.Order{}, err
	}

	//builder.SetEnvironment(env.Live())

	// Step 1: Retrieve membership for this user.
	// The membership might be empty but the value is
	// valid.
	sugar.Infof("Start retrieving membership for reader %+v", builder.GetReaderID())
	member, err := otx.RetrieveMember(builder.GetReaderID())
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.Order{}, err
	}
	sugar.Infof("Membership retrieved %+v", member)

	err = builder.DeduceSubsKind(member)
	if err != nil {
		sugar.Error()
		_ = otx.Rollback()
		return subs.Order{}, err
	}

	sugar.Infof("Subscription kind %s", builder.GetSubsKind())

	// Step 2: Build an order for the user's chosen plan
	// with chosen payment method based on previous
	// membership so that we could how this order
	// is used: create, renew or upgrade.

	// Step 3: required only if this order is used for
	// upgrading.
	if builder.GetSubsKind() == enum.OrderKindUpgrade {
		// Step 3.1: find previous orders with balance
		// remaining.
		// DO not save sources directly. The balance is not
		// calculated at this point.
		sugar.Infof("Get balance sources for an upgrading order")
		orders, err := otx.FindBalanceSources(builder.GetReaderID())
		if err != nil {
			sugar.Error(err)
			_ = otx.Rollback()
			return subs.Order{}, err
		}
		sugar.Infof("Find prorated orders: %+v", orders)

		// Step 3.2: Build wallet
		wallet := subs.NewWallet(orders, time.Now())

		builder.SetWallet(wallet)
	}

	// Now all data are collected. Build order.
	if err := builder.Build(); err != nil {
		_ = otx.Rollback()
		return subs.Order{}, err
	}

	order, err := builder.GetOrder()
	if err != nil {
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
	if order.Kind == enum.OrderKindUpgrade {

		if err := otx.SaveProratedOrders(builder.GetWallet().Sources); err != nil {
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
