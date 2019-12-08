package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/plan"
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

// TODO: change to UpgradeSchema
func (router UpgradeRouter) UpgradeBalance(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	otx, err := router.subEnv.BeginOrderTx()
	if err != nil {
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	member, err := otx.RetrieveMember(userID)
	if err != nil {
		_ = otx.Rollback()
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	orders, err := otx.FindBalanceSources(userID)
	if err != nil {
		_ = otx.Rollback()
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	if err := otx.Commit(); err != nil {
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	wallet := subscription.NewWallet(orders, time.Now())
	p, _ := plan.FindPlan(enum.TierPremium, enum.CycleYear)

	builder := subscription.NewOrderBuilder(userID).
		SetPlan(p).
		SetMembership(member).
		SetWallet(wallet)

	if router.subEnv.UseSandbox() {
		builder.SetSandbox()
	}

	pi, _ := builder.PaymentIntent()

	switch pi.SubsKind {
	case plan.SubsKindNull:
		_ = view.Render(w, view.NewBadRequest("Cannot determine your current subscription status"))
		return

	case plan.SubsKindCreate:
		_ = view.Render(w, view.NewBadRequest("No a valid member yet."))
		return

	case plan.SubsKindRenew:
		_ = view.Render(w, view.NewBadRequest("Not used for renewal."))
		return
	}

	if !member.IsAliOrWxPay() {
		_ = view.Render(w, view.NewBadRequest("Only payment made via alipay or wechat pay has balance"))
		return
	}

	_ = view.Render(w, view.NewResponse().SetBody(pi))
}

func (router UpgradeRouter) FreeUpgrade(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	p, _ := plan.FindPlan(enum.TierPremium, enum.CycleYear)
	clientApp := util.NewClientApp(req)

	builder := subscription.NewOrderBuilder(userID).
		SetPlan(p).
		SetClient(clientApp)

	if router.subEnv.UseSandbox() {
		builder.SetSandbox()
	}

	otx, err := router.subEnv.BeginOrderTx()
	if err != nil {
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	member, err := otx.RetrieveMember(userID)
	if err != nil {
		_ = otx.Rollback()
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	orders, err := otx.FindBalanceSources(userID)
	if err != nil {
		_ = otx.Rollback()
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	wallet := subscription.NewWallet(orders, time.Now())
	builder.SetMembership(member).
		SetWallet(wallet)

	pi, _ := builder.PaymentIntent()

	switch pi.SubsKind {
	case plan.SubsKindNull:
		_  = otx.Rollback()
		_ = view.Render(w, view.NewBadRequest("Cannot determine your current subscription status"))
		return

	case plan.SubsKindCreate:
		_ = otx.Rollback()
		_ = view.Render(w, view.NewBadRequest("No a valid member yet."))
		return

	case plan.SubsKindRenew:
		_ = otx.Rollback()
		_ = view.Render(w, view.NewBadRequest("Not used for renewal."))
		return
	}

	if !member.IsAliOrWxPay() {
		_ = otx.Rollback()
		_ = view.Render(w, view.NewBadRequest("Only payment made via alipay or wechat pay has balance"))
		return
	}

	// If user needs to pay any extra money.
	if pi.Plan.Amount > pi.Wallet.Balance {
		_ = view.Render(w, view.NewResponse().SetBody(pi))
		return
	}

	upgradeSchema, _ := builder.UpgradeSchema()
	if err := otx.SaveUpgradeIntent(upgradeSchema); err != nil {
		_ = otx.Rollback()
		return
	}

	if err := otx.SaveProratedOrders(builder.ProratedOrdersSchema()); err != nil {

		_ = otx.Rollback()

		return
	}

	order, err := builder.Order()

	if err != nil {
		_ = otx.Rollback()
		return
	}

	snapshot := subscription.NewMemberSnapshot(
		member,
		plan.SubsKindUpgrade.SnapshotReason(),
	)

	order.MemberSnapshotID = null.StringFrom(snapshot.SnapshotID)

	if err := otx.SaveOrder(order); err != nil {
		_ = otx.Rollback()
		return
	}

	newMember, err := member.FromAliOrWx(order)

	if err := otx.UpdateMember(newMember); err != nil {
		_ = otx.Rollback()
		return
	}

	if err := otx.Commit(); err != nil {
		return
	}

	// Save client app info
	go func() {
		_ = router.subEnv.SaveOrderClient(builder.ClientApp())
	}()

	go func() {
		_ = router.subEnv.BackUpMember(snapshot)
	}()

	go func() {
		err := router.sendConfirmationEmail(order)
		if err != nil {
			logrus.WithField("trace", "UpgradeRouter.FreeUpgrade").Error(err)
		}
	}()

	_ = view.Render(w, view.NewNoContent())
}

