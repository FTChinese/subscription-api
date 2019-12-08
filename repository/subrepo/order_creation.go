package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"time"
)

func (env SubEnv) CreateOrder(builder *subscription.OrderBuilder) (subscription.Order, error) {
	log := logger.WithField("trace", "PayRouter.createOrder")

	otx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return subscription.Order{}, err
	}

	if env.UseSandbox() {
		builder.SetSandbox()
	}

	// Step 1: Retrieve membership for this user.
	// The membership might be empty but the value is
	// valid.
	log.Infof("Start retrieving membership for reader %+v", builder.GetReaderID())
	member, err := otx.RetrieveMember(builder.GetReaderID())
	if err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return subscription.Order{}, err
	}
	log.Infof("Membership retrieved %+v", member)

	// Optional: add member id is this member exists but
	// its id field is empty.
	if !member.IsZero() && member.ID.IsZero() {

		member.GenerateID()
		log.Infof("Membership does not have an id. Generated and add it %s", member.ID.String)

		go func() {
			if err := env.AddMemberID(member); err != nil {
				log.Error(err)
			}
		}()
	}

	builder.SetMembership(member)
	subKind, err := builder.GetSubsKind()
	if err != nil {
		_ = otx.Rollback()
		return subscription.Order{}, err
	}

	log.Infof("Subscription kind %s", subKind)

	// Step 2: Build an order for the user's chosen plan
	// with chosen payment method based on previous
	// membership so that we could how this order
	// is used: create, renew or upgrade.

	// Step 3: required only if this order is used for
	// upgrading.
	if subKind == plan.SubsKindUpgrade {
		// Step 3.1: find previous orders with balance
		// remaining.
		// DO not save sources directly. The balance is not
		// calculated at this point.
		log.Infof("Get balance sources for an upgrading order")
		orders, err := otx.FindBalanceSources(builder.GetReaderID())
		if err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return subscription.Order{}, err
		}
		log.Infof("Find prorated orders: %+v", orders)

		// Step 3.2: Build upgrade plan
		wallet := subscription.NewWallet(orders, time.Now())

		builder.SetWallet(wallet)
	}

	if err := builder.Build(); err != nil {
		_ = otx.Rollback()
		return subscription.Order{}, err
	}

	order, err := builder.Order()
	if err != nil {
		return subscription.Order{}, err
	}

	if subKind == plan.SubsKindUpgrade {
		upIntent, _ := builder.UpgradeSchema()
		// Step 3.4: Save the upgrade plan
		if err := otx.SaveUpgradeIntent(upIntent); err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return subscription.Order{}, err
		}
		log.Info("Upgrade intent saved")

		// Step 3.5: Save prorated orders
		if err := otx.SaveProratedOrders(builder.ProratedOrdersSchema()); err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return subscription.Order{}, err
		}
		log.Info("Prorated orders saved")
	}

	// Step 4: Save this order.
	if err := otx.SaveOrder(order); err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return subscription.Order{}, err
	}
	log.Infof("Order saved %s", order.ID)

	if err := otx.Commit(); err != nil {
		log.Error(err)
		return subscription.Order{}, err
	}

	return order, nil
}