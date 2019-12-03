package controller

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/subrepo"
	"net/http"
	"time"
)

type UpgradeRouter struct {
	PayRouter
}

func NewUpgradeRouter(env subrepo.SubEnv) UpgradeRouter {
	r := UpgradeRouter{}
	r.subEnv = env

	return r
}

func (router UpgradeRouter) getUpgradePlan(id reader.MemberID) (subscription.UpgradeIntent, error) {
	log := logrus.WithField("trace", "UpgradeRouter.getUpgradePlan")

	otx, err := router.subEnv.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return subscription.UpgradeIntent{}, err
	}

	member, err := otx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return subscription.UpgradeIntent{}, err
	}

	// To upgrade, membership must exist, not expired yet,
	// must be alipay or wxpay, and must not be premium.
	if member.IsZero() {
		_ = otx.Rollback()
		return subscription.UpgradeIntent{}, sql.ErrNoRows
	}

	if member.IsExpired() {
		_ = otx.Rollback()
		return subscription.UpgradeIntent{}, util.ErrMemberExpired
	}

	if member.PaymentMethod == enum.PayMethodStripe {
		_ = otx.Rollback()
		return subscription.UpgradeIntent{}, util.ErrValidStripeSwitching
	}

	if member.Tier == enum.TierPremium {
		_ = otx.Rollback()
		return subscription.UpgradeIntent{}, util.ErrAlreadyUpgraded
	}

	orders, err := otx.FindBalanceSources(id)
	if err != nil {
		_ = otx.Rollback()
		return subscription.UpgradeIntent{}, err
	}

	if err := otx.Commit(); err != nil {
		log.Error(err)
		return subscription.UpgradeIntent{}, err
	}

	wallet := subscription.NewWallet(orders, time.Now())

	premiumPlan, _ := plan.FindFtcPlan("premium_year")

	up := subscription.NewUpgradeIntent(wallet, premiumPlan)

	return up, nil
}

// TODO: change to UpgradeIntent
func (router UpgradeRouter) UpgradeBalance(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	up, err := router.getUpgradePlan(userID)
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	_ = view.Render(w, view.NewResponse().SetBody(up))
}

func (router UpgradeRouter) FreeUpgrade(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	up, err := router.getUpgradePlan(userID)
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// If user needs to pay any extra money.
	if up.Plan.NetPrice > 0 {
		_ = view.Render(w, view.NewResponse().SetBody(up))
		return
	}

	subs, err := router.freeUpgrade(userID, util.NewClientApp(req))
	if err != nil {
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	go func() {
		err := router.sendConfirmationEmail(subs)
		if err != nil {
			logrus.WithField("trace", "UpgradeRouter.FreeUpgrade").Error(err)
		}
	}()

	_ = view.Render(w, view.NewNoContent())
}

func (router UpgradeRouter) freeUpgrade(id reader.MemberID, app util.ClientApp) (subscription.Order, error) {
	log := logrus.WithField("trace", "UpgradeRouter.freeUpgrade")

	tx, err := router.subEnv.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return subscription.Order{}, err
	}

	member, err := tx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return subscription.Order{}, err
	}

	// To upgrade, membership must exist, not expired yet,
	// must be alipay or wxpay, and must not be premium.
	if member.IsZero() {
		_ = tx.Rollback()
		return subscription.Order{}, sql.ErrNoRows
	}

	if member.IsExpired() {
		_ = tx.Rollback()
		return subscription.Order{}, util.ErrMemberExpired
	}

	if member.PaymentMethod == enum.PayMethodStripe {
		_ = tx.Rollback()
		return subscription.Order{}, util.ErrValidStripeSwitching
	}

	if member.Tier == enum.TierPremium {
		_ = tx.Rollback()
		return subscription.Order{}, util.ErrAlreadyUpgraded
	}

	orders, err := tx.FindBalanceSources(id)
	if err != nil {
		_ = tx.Rollback()
		return subscription.Order{}, err
	}

	wallet := subscription.NewWallet(orders, time.Now())
	premiumPlan, _ := plan.FindFtcPlan("premium_year")
	up := subscription.NewUpgradeIntent(wallet, premiumPlan)
	if up.Plan.NetPrice > 0 {
		return subscription.Order{}, errors.New("you cannot upgrade for free since payment is required")
	}

	if err := tx.SaveUpgradeIntent(up); err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return subscription.Order{}, err
	}

	if err := tx.SaveProratedOrders(up.Data); err != nil {
		log.Error(err)
		_ = tx.Rollback()

		return subscription.Order{}, err
	}

	order, err := subscription.NewFreeUpgradeOrder(id, up)

	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return order, err
	}

	snapshot := subscription.NewMemberSnapshot(member, subscription.SubsKindUpgrade.SnapshotReason())
	order.MemberSnapshotID = null.StringFrom(snapshot.SnapshotID)

	if err := tx.SaveOrder(order); err != nil {
		_ = tx.Rollback()
		return order, err
	}

	newMember, err := member.FromAliOrWx(order)

	if err := tx.UpdateMember(newMember); err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return order, err
	}

	if err := tx.Commit(); err != nil {
		log.Error(err)
		return order, err
	}

	// Save client app info
	go func() {
		if err := router.subEnv.SaveOrderClient(order.ID, app); err != nil {
			log.Error(err)
		}
	}()

	go func() {
		if err := router.subEnv.BackUpMember(snapshot); err != nil {
			log.Error()
		}
	}()

	return order, nil
}
