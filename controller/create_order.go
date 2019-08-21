package controller

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/util"
)

// createOrder creates an order for ali or wx pay.
func (router PayRouter) createOrder(
	id paywall.AccountID,
	plan paywall.Plan,
	method enum.PayMethod,
	app util.ClientApp,
	wxAppId null.String,
) (paywall.Subscription, error) {
	log := logrus.WithField("trace", "PayRouter.createOrder")

	if method != enum.PayMethodWx && method != enum.PayMethodAli {
		return paywall.Subscription{}, errors.New("only used by alipay or wxpay")
	}

	otx, err := router.env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.Subscription{}, err
	}

	log.Infof("Start retrieving membership for reader %+v", id)
	member, err := otx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return paywall.Subscription{}, err
	}
	log.Infof("Membership retrieved %+v", member)

	order, err := paywall.NewOrder(id, plan, method, member)
	if err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return paywall.Subscription{}, err
	}

	// If this order is used for upgrading
	if order.Usage == paywall.SubsKindUpgrade {
		balanceSource, err := otx.FindBalanceSources(id)
		if err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return paywall.Subscription{}, err
		}
		log.Infof("Find balance source: %+v", balanceSource)

		var up paywall.UpgradePreview
		order, up = order.WithUpgrade(balanceSource)

		if err := otx.SaveUpgradeV2(up, member); err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return paywall.Subscription{}, err
		}

		if err := otx.SetLastUpgradeIDV2(up); err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return paywall.Subscription{}, err
		}
	}

	order.WxAppID = wxAppId

	if err := otx.SaveOrder(order, app); err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return paywall.Subscription{}, err
	}

	if err := otx.Commit(); err != nil {
		log.Error(err)
		return paywall.Subscription{}, err
	}

	if !router.env.Live() {
		order.Amount = 0.01
	}

	return order, nil
}
