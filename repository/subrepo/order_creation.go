package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"time"
)

func (env Env) CreateOrder(builder *subs.OrderBuilder) (subs.Order, error) {
	log := logger.WithField("trace", "PayRouter.createOrder")

	otx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return subs.Order{}, err
	}

	//builder.SetEnvironment(env.Live())

	// Step 1: Retrieve membership for this user.
	// The membership might be empty but the value is
	// valid.
	log.Infof("Start retrieving membership for reader %+v", builder.GetReaderID())
	member, err := otx.RetrieveMember(builder.GetReaderID())
	if err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return subs.Order{}, err
	}
	log.Infof("Membership retrieved %+v", member)

	err = builder.DeduceSubsKind(member)
	if err != nil {
		_ = otx.Rollback()
		return subs.Order{}, err
	}

	log.Infof("Subscription kind %s", builder.GetSubsKind())

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
		log.Infof("Get balance sources for an upgrading order")
		orders, err := otx.FindBalanceSources(builder.GetReaderID())
		if err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return subs.Order{}, err
		}
		log.Infof("Find prorated orders: %+v", orders)

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
		log.Error(err)
		_ = otx.Rollback()
		return subs.Order{}, err
	}
	log.Infof("Order saved %s", order.ID)

	// Step 5: Save prorated orders for upgrade.
	if order.Kind == enum.OrderKindUpgrade {

		if err := otx.SaveProratedOrders(builder.GetWallet().Sources); err != nil {
			_ = otx.Rollback()
			return subs.Order{}, err
		}
	}

	if err := otx.Commit(); err != nil {
		log.Error(err)
		return subs.Order{}, err
	}

	return order, nil
}
