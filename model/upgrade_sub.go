package model

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
	"strings"
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
	up.Member = member

	return up, nil
}

func (env Env) FreeUpgrade(
	userID paywall.UserID,
	up paywall.UpgradePreview,
	clientApp util.ClientApp,
) (paywall.Subscription, error) {

	log := logger.WithField("trace", "Env.PreviewUpgrade")

	subs, err := paywall.NewSubs(userID, up.Plan)
	if err != nil {
		return subs, err
	}
	subs.Usage = paywall.SubsKindUpgrade

	tx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return subs, err
	}

	if err := tx.SaveOrder(subs, clientApp); err != nil {
		_ = tx.rollback()
		return subs, err
	}

	if err := tx.SaveUpgradeV2(subs.ID, up); err != nil {
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

	return env.ConfirmPayment(subs.ID, time.Now())
}

// LoadUpgradeSource retrieves upgrading balance and sources from an upgrade order.
func (env Env) LoadUpgradeSource(orderID string) (paywall.UpgradePreview, error) {
	var up paywall.UpgradePreview
	var sources string
	err := env.db.QueryRow(env.query.SelectUpgrade(), orderID).Scan(
		&up.ID,
		&up.Balance,
		&sources,
		&up.OrderID,
		&up.CreatedAt,
		&up.ConfirmedAt)

	up.SourceIDs = strings.Split(sources, ",")
	if err != nil {
		return up, err
	}

	return up, nil
}

// DirectUpgradeOrder creates an order for upgrading without
// requiring user to pay. This almost will never be used since
// user must have enough balance to cover upgrading cost,
// which nearly won't happen since we limit renewal to 3 years
// at most. 3 years of standard membership costs 258 * 3 < 1998.
// It is provided here just for completeness.
// Deprecate
func (env Env) DirectUpgradeOrder(
	user paywall.UserID,
	upgrade paywall.Upgrade,
	clientApp util.ClientApp) (paywall.Subscription, error) {

	subs, err := paywall.NewUpgradeOrder(user, upgrade)
	if err != nil {
		return subs, err
	}

	otx, err := env.BeginOrderTx()
	if err != nil {
		logger.WithField("trace", "Env.CreateOrder").Error(err)
		return paywall.Subscription{}, err
	}

	if err = otx.SaveOrder(subs, clientApp); err != nil {
		_ = otx.rollback()
		return subs, err
	}

	if err = otx.SaveUpgrade(subs.ID, upgrade); err != nil {
		_ = otx.rollback()
		return subs, err
	}

	if err := otx.SetUpgradeIDOnSource(upgrade); err != nil {
		_ = otx.rollback()
		return subs, err
	}

	if err := otx.commit(); err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return paywall.Subscription{}, err
	}

	// Return the order
	return subs, nil
}

// Upgrade builds upgrade preview for a standard user who
// is trying to upgrade to premium.
// DO remember to rollback!
// Deprecate
func (env Env) UpgradeBalance(user paywall.UserID) (paywall.Upgrade, error) {
	otx, err := env.BeginOrderTx()
	if err != nil {
		logger.WithField("trace", "Env.CreateOrder").Error(err)
		return paywall.Upgrade{}, err
	}

	member, err := otx.RetrieveMember(user)
	// If membership is not found for this user, deny upgrading.
	if err != nil {
		_ = otx.rollback()
		return paywall.Upgrade{}, err
	}

	if member.IsZero() {
		_ = otx.rollback()
		return paywall.Upgrade{}, util.ErrMemberNotFound
	}

	if member.Tier == enum.TierPremium {
		_ = otx.rollback()
		return paywall.Upgrade{}, util.ErrAlreadyUpgraded
	}

	orders, err := otx.FindBalanceSources(user)
	if err != nil {
		_ = otx.rollback()
		return paywall.Upgrade{}, err
	}

	plan, err := paywall.GetFtcPlans(true).FindPlan("premium_year")
	if err != nil {
		return paywall.Upgrade{}, err
	}
	up := paywall.NewUpgrade(plan).
		SetBalance(orders).
		CalculatePayable()

	if err := otx.commit(); err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return up, err
	}

	return up, nil
}
