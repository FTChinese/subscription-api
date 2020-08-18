package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/builder"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"time"
)

func (env SubEnv) CreateOrder(builder *builder.OrderBuilder) (subs.Order, error) {
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
	// TODO: changed sql where clause.
	member, err := otx.RetrieveMember(builder.GetReaderID())
	if err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return subs.Order{}, err
	}
	log.Infof("Membership retrieved %+v", member)

	builder.SetMembership(member)
	subKind, err := builder.GetSubsKind()
	if err != nil {
		_ = otx.Rollback()
		return subs.Order{}, err
	}

	log.Infof("Subscription kind %s", subKind)

	// Step 2: Build an order for the user's chosen plan
	// with chosen payment method based on previous
	// membership so that we could how this order
	// is used: create, renew or upgrade.

	// Step 3: required only if this order is used for
	// upgrading.
	if subKind == enum.OrderKindUpgrade {
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

		// Step 3.2: Build upgrade plan
		wallet := subs.NewWallet(orders, time.Now())

		builder.SetWallet(wallet)
	}

	if err := builder.Build(); err != nil {
		_ = otx.Rollback()
		return subs.Order{}, err
	}

	order, err := builder.Order()
	if err != nil {
		return subs.Order{}, err
	}

	if subKind == enum.OrderKindUpgrade {
		upgrade, _ := builder.UpgradeSchema()

		// Step 3.4: Save the upgrade plan
		// Step 3.5: Save prorated orders
		if err := otx.SaveUpgradeSchema(upgrade); err != nil {
			_ = otx.Rollback()
			return subs.Order{}, err
		}
	}

	// Step 4: Save this order.
	if err := otx.SaveOrder(order); err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return subs.Order{}, err
	}
	log.Infof("Order saved %s", order.ID)

	if err := otx.Commit(); err != nil {
		log.Error(err)
		return subs.Order{}, err
	}

	return order, nil
}
