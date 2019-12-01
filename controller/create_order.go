package controller

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/models/util"
)

// createOrder creates an order for ali or wx pay.
func (router PayRouter) createOrder(
	id reader.MemberID,
	p plan.Plan,
	method enum.PayMethod,
	app util.ClientApp,
	wxAppId null.String,
) (subscription.Order, error) {
	log := logrus.WithField("trace", "PayRouter.createOrder")

	if method != enum.PayMethodWx && method != enum.PayMethodAli {
		return subscription.Order{}, errors.New("only used by alipay or wxpay")
	}

	otx, err := router.env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return subscription.Order{}, err
	}

	// Step 1: Retrieve membership for this user.
	// The membership might be empty but the value is
	// valid.
	log.Infof("Start retrieving membership for reader %+v", id)
	member, err := otx.RetrieveMember(id)
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
			if err := router.env.AddMemberID(member); err != nil {
				log.Error(err)
			}
		}()
	}

	// Step 2: Build an order for the user's chosen plan
	// with chosen payment method based on previous
	// membership so that we could how this order
	// is used: create, renew or upgrade.
	order, err := subscription.NewOrder(id, p, method, member)
	if err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return subscription.Order{}, err
	}

	log.Infof("Created an order %s for %s", order.ID, order.Usage)

	// Step 3: required only if this order is used for
	// upgrading.
	if order.Usage == subscription.SubsKindUpgrade {
		// Step 3.1: find previous orders with balance
		// remaining.
		// DO not save sources directly. The balance is not
		// calculated at this point.
		log.Infof("Get balance sources for an upgrading order")
		sources, err := otx.FindBalanceSources(id)
		if err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return subscription.Order{}, err
		}
		log.Infof("Find balance source: %+v", sources)

		// Step 3.2: Build upgrade plan
		up := plan.NewUpgradePlan(sources)
		log.Infof("Upgrading plan: %+v", up)

		// Step 3.3: Update order based on upgrade plan.
		order = order.WithUpgrade(up)

		// Step 3.4: Save the upgrade plan
		if err := otx.SaveUpgradePlan(up); err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return subscription.Order{}, err
		}
		log.Info("Upgrading plan saved")

		// Step 3.5: Save prorated orders
		if err := otx.SaveProration(up.Data); err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return subscription.Order{}, err
		}
		log.Info("Prorated orders saved")
	}

	order.WxAppID = wxAppId

	// Back up membership state the moment the order is created.
	if !member.IsZero() {

		snapshot := subscription.NewMemberSnapshot(member, order.Usage)

		order.MemberSnapshotID = null.StringFrom(snapshot.SnapshotID)

		log.Infof("Membership is not empty. Take a snapshot of its current status %s", snapshot.SnapshotID)

		go func() {
			if err := router.env.BackUpMember(snapshot); err != nil {
				log.Error(err)
			}
		}()
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

	// Not vital. Perform in background.
	go func() {
		if err := router.env.SaveOrderClient(order.ID, app); err != nil {
			log.Error(err)
		}
	}()

	if !router.env.Live() {
		order.Amount = 0.01
	}

	return order, nil
}
