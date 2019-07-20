package model

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
	"time"
)

func (env Env) PreviewUpgrade(userID paywall.UserID) (paywall.UpgradePreview, error) {

	log := logger.WithField("trace", "Env.PreviewUpgrade")

	otx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.UpgradePreview{}, err
	}

	member, err := otx.RetrieveMember(userID)
	if err != nil {
		_ = otx.rollback()
		return paywall.UpgradePreview{}, err
	}

	if member.IsZero() {
		_ = otx.rollback()
		return paywall.UpgradePreview{}, util.ErrMemberNotFound
	}

	if member.IsExpired() {
		return paywall.UpgradePreview{}, errors.New("please re-subscribe to any product since your subscription is already expired")
	}

	if member.PaymentMethod == enum.PayMethodStripe {
		_ = otx.rollback()
		return paywall.UpgradePreview{}, errors.New("stripe user cannot switch plan via alipay or wechat pay")
	}

	if member.Tier == enum.TierPremium {
		_ = otx.rollback()
		return paywall.UpgradePreview{}, util.ErrAlreadyUpgraded
	}

	sources, err := otx.FindBalanceSources(userID)
	if err != nil {
		_ = otx.rollback()
		return paywall.UpgradePreview{}, err
	}

	if err := otx.commit(); err != nil {
		log.Error(err)
		return paywall.UpgradePreview{}, err
	}

	up := paywall.NewUpgradePreview(sources)
	plan, _ := env.GetCurrentPlans().GetPlanByID("premium_year")
	up.Plan = plan.BuildUpgradePlan(up.Balance)
	up.Member = member

	return up, nil
}

func (env Env) FreeUpgrade(
	userID paywall.UserID,
	up paywall.UpgradePreview,
	clientApp util.ClientApp,
) (paywall.Subscription, error) {

	log := logger.WithField("trace", "Env.PreviewUpgrade")

	subs, err := paywall.NewUpgradeOrderV2(userID, up)
	if err != nil {
		return subs, err
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return subs, err
	}

	if err := tx.SaveOrder(subs, clientApp); err != nil {
		_ = tx.rollback()
		return subs, err
	}

	if err := tx.SaveUpgradeV2(subs.OrderID, up); err != nil {
		_ = tx.rollback()
		return subs, err
	}

	if err := tx.SetUpgradeIDOnSourceV2(up); err != nil {
		_ = tx.rollback()
		return subs, err
	}

	if err := tx.commit(); err != nil {
		log.Error(err)
		return subs, err
	}

	return env.ConfirmPayment(subs.OrderID, time.Now())
}
