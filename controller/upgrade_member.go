package controller

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository"
	"net/http"
)

type UpgradeRouter struct {
	PayRouter
}

func NewUpgradeRouter(env repository.Env) UpgradeRouter {
	r := UpgradeRouter{}
	r.env = env

	return r
}

func (router UpgradeRouter) getUpgradePlan(id reader.AccountID) (paywall.UpgradePlan, error) {
	log := logrus.WithField("trace", "UpgradeRouter.getUpgradePlan")

	otx, err := router.env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.UpgradePlan{}, err
	}

	member, err := otx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return paywall.UpgradePlan{}, err
	}

	// To upgrade, membership must exist, not expired yet,
	// must be alipay or wxpay, and must not be premium.
	if member.IsZero() {
		_ = otx.Rollback()
		return paywall.UpgradePlan{}, sql.ErrNoRows
	}

	if member.IsExpired() {
		_ = otx.Rollback()
		return paywall.UpgradePlan{}, util.ErrMemberExpired
	}

	if member.PaymentMethod == enum.PayMethodStripe {
		_ = otx.Rollback()
		return paywall.UpgradePlan{}, util.ErrValidStripeSwitching
	}

	if member.Tier == enum.TierPremium {
		_ = otx.Rollback()
		return paywall.UpgradePlan{}, util.ErrAlreadyUpgraded
	}

	sources, err := otx.FindBalanceSources(id)
	if err != nil {
		_ = otx.Rollback()
		return paywall.UpgradePlan{}, err
	}

	if err := otx.Commit(); err != nil {
		log.Error(err)
		return paywall.UpgradePlan{}, err
	}

	up := paywall.NewUpgradePlan(sources)

	return up, nil
}

func (router UpgradeRouter) UpgradeBalance(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	up, err := router.getUpgradePlan(userID)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	view.Render(w, view.NewResponse().SetBody(up))
}

func (router UpgradeRouter) FreeUpgrade(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	up, err := router.getUpgradePlan(userID)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// If user needs to pay any extra money.
	if up.Plan.NetPrice > 0 {
		view.Render(w, view.NewResponse().SetBody(up))
		return
	}

	subs, err := router.freeUpgrade(userID, util.NewClientApp(req))
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	go func() {
		err := router.sendConfirmationEmail(subs)
		if err != nil {
			logrus.WithField("trace", "UpgradeRouter.FreeUpgrade").Error(err)
		}
	}()

	view.Render(w, view.NewNoContent())
}

func (router UpgradeRouter) freeUpgrade(id reader.AccountID, app util.ClientApp) (paywall.Order, error) {
	log := logrus.WithField("trace", "UpgradeRouter.freeUpgrade")

	tx, err := router.env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.Order{}, err
	}

	member, err := tx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return paywall.Order{}, err
	}

	// To upgrade, membership must exist, not expired yet,
	// must be alipay or wxpay, and must not be premium.
	if member.IsZero() {
		_ = tx.Rollback()
		return paywall.Order{}, sql.ErrNoRows
	}

	if member.IsExpired() {
		_ = tx.Rollback()
		return paywall.Order{}, util.ErrMemberExpired
	}

	if member.PaymentMethod == enum.PayMethodStripe {
		_ = tx.Rollback()
		return paywall.Order{}, util.ErrValidStripeSwitching
	}

	if member.Tier == enum.TierPremium {
		_ = tx.Rollback()
		return paywall.Order{}, util.ErrAlreadyUpgraded
	}

	sources, err := tx.FindBalanceSources(id)
	if err != nil {
		_ = tx.Rollback()
		return paywall.Order{}, err
	}

	up := paywall.NewUpgradePlan(sources)
	if up.Plan.NetPrice > 0 {
		return paywall.Order{}, errors.New("you cannot upgrade for free since payment is required")
	}

	if err := tx.SaveUpgradePlan(up); err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return paywall.Order{}, err
	}

	if err := tx.SaveProration(up.Data); err != nil {
		log.Error(err)
		_ = tx.Rollback()

		return paywall.Order{}, err
	}

	order, err := paywall.NewFreeUpgradeOrder(id, up)

	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return order, err
	}

	snapshot := paywall.NewMemberSnapshot(member, paywall.SubsKindUpgrade)
	order.MemberSnapshotID = null.StringFrom(snapshot.ID)

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
		if err := router.env.SaveOrderClient(order.ID, app); err != nil {
			log.Error(err)
		}
	}()

	go func() {
		if err := router.env.BackUpMember(snapshot); err != nil {
			log.Error()
		}
	}()

	return order, nil
}
