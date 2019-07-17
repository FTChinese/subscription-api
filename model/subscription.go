package model

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/util"
	"time"

	"gitlab.com/ftchinese/subscription-api/paywall"
)

// CreateOrder creates an order for a user with the selected plan.
// Only payment method alipay and wechat pay is allowed.
func (env Env) CreateOrder(
	user paywall.UserID,
	plan paywall.Plan,
	payMethod enum.PayMethod,
	clientApp util.ClientApp,
	wxAppId null.String,
) (paywall.Subscription, error) {

	if payMethod != enum.PayMethodWx && payMethod != enum.PayMethodAli {
		return paywall.Subscription{}, errors.New("only alipay and wechat pay are allowed here")
	}

	log := logger.WithField("trace", "Env.CreateOrder")

	otx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.Subscription{}, err
	}

	// Find the member
	log.Info("Start retrieving membership")
	member, err := otx.RetrieveMember(user)
	if err != nil {
		_ = otx.rollback()
		return paywall.Subscription{}, err
	}

	log.Infof("Membership retrieved %+v", member)

	// UserID the found member to determine order category.
	// Ignore Golang warning here. It is safe here to user
	// the zero value of membership since it is not a pointer.
	subsKind, err := member.SubsKind(plan)
	if err != nil {
		return paywall.Subscription{}, err
	}

	log.Infof("Order used for %s", subsKind)

	// Create order
	var subs paywall.Subscription
	if subsKind == paywall.SubsKindUpgrade {
		// For upgrade order, first find user's balance.
		balanceSource, err := otx.FindBalanceSources(user)
		if err != nil {
			_ = otx.rollback()
			return subs, err
		}
		log.Infof("Find balance source: %+v", balanceSource)

		upgrade := paywall.NewUpgrade(plan).
			SetBalance(balanceSource).
			CalculatePayable()

		log.Infof("Upgrading scheme: %+v", upgrade)

		// Record the membership status prior to upgrade.
		upgrade.Member = member
		subs, err = paywall.NewUpgradeOrder(user, upgrade)
		if err != nil {
			_ = otx.rollback()
			return subs, err
		}

		err = otx.SaveUpgrade(subs.OrderID, upgrade)
		if err != nil {
			_ = otx.rollback()
			return subs, err
		}

		log.Infof("Saved upgrade scheme %s", upgrade.ID)

		err = otx.SetUpgradeIDOnSource(upgrade)
		if err != nil {
			_ = otx.rollback()
			return subs, err
		}

		log.Info("Set upgrade scheme id on balance source")

	} else {
		subs, err = paywall.NewSubs(user, plan)
		if err != nil {
			return subs, err
		}
		subs.Usage = subsKind
	}
	subs.PaymentMethod = payMethod
	subs.WxAppID = wxAppId

	log.Infof("Subscription: %+v", subs)
	err = otx.SaveOrder(subs, clientApp)
	if err != nil {
		_ = otx.rollback()
		return subs, err
	}

	if err := otx.commit(); err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return paywall.Subscription{}, err
	}

	// For sandbox.
	if env.sandbox {
		log.Info("Sandbox environment. Set price to 0.01")
		subs.NetPrice = 0.01
		return subs, nil
	}
	// Return the order
	return subs, nil
}

// Upgrade builds upgrade preview for a standard user who
// is trying to upgrade to premium.
// DO remember to rollback!
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
		return paywall.Upgrade{}, ErrMemberNotFound
	}

	if member.Tier == enum.TierPremium {
		_ = otx.rollback()
		return paywall.Upgrade{}, ErrAlreadyUpgraded
	}

	orders, err := otx.FindBalanceSources(user)
	if err != nil {
		_ = otx.rollback()
		return paywall.Upgrade{}, err
	}

	plan, _ := env.GetCurrentPlans().FindPlan(enum.TierPremium.String(), enum.CycleYear.String())

	up := paywall.NewUpgrade(plan).
		SetBalance(orders).
		CalculatePayable()

	if err := otx.commit(); err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return up, err
	}

	// For testing.
	// We must mimic the whole process.
	if env.sandbox {
		up.Payable = 0.01
		return up, nil
	}

	return up, nil
}

// DirectUpgradeOrder creates an order for upgrading without
// requiring user to pay. This almost will never be used since
// user must have enough balance to cover upgrading cost,
// which nearly won't happen since we limit renewal to 3 years
// at most. 3 years of standard membership costs 258 * 3 < 1998.
// It is provided here just for completeness.
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

	// Save Order
	if env.sandbox {
		subs.NetPrice = 0.01
	}

	if err = otx.SaveOrder(subs, clientApp); err != nil {
		_ = otx.rollback()
		return subs, err
	}

	if err = otx.SaveUpgrade(subs.OrderID, upgrade); err != nil {
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

// FindSubscription tries to find an order to verify the authenticity of a subscription order.
func (env Env) FindSubsCharge(orderID string) (paywall.Charge, error) {

	var c paywall.Charge
	err := env.db.QueryRow(
		env.query.SelectSubsPrice(),
		orderID,
	).Scan(
		&c.ListPrice,
		&c.NetPrice,
		&c.IsConfirmed,
	)

	if err != nil {
		logger.WithField("trace", "Env.FindSubsCharge").Error(err)
		return c, err
	}

	return c, nil
}

// ConfirmPayment handles payment notification with database locking.
// Returns the a complete Subscription to be used to compose an email.
// If returned error is ErrOrderNotFound or ErrAlreadyConfirmed, tell Wechat or Ali do not try any more; otherwise let them retry.
// Only when error is nil should be send a confirmation email.
// States passed back:
// Error occurred, allow retry;
// Error occurred, don't retry;
// No error, send user confirmation letter.
// Concurrency pitfalls: if a user, whose is not a member yet, paid at the same moment twice, there are chances that those two orders are both used to create a membership, since transaction lock for update works only when a row exists.
func (env Env) ConfirmPayment(orderID string, confirmedAt time.Time) (paywall.Subscription, error) {
	log := logger.WithField("trace", "Env.ConfirmPayment")

	tx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.Subscription{}, ErrAllowRetry
	}

	// Step 1: Find the subscription order by order id
	// The row is locked for update.
	// If the order is not found, or is already confirmed,
	// tell provider not sending notification any longer;
	// otherwise, allow retry.
	subs, errSubs := tx.RetrieveOrder(orderID)
	if errSubs != nil {
		_ = tx.rollback()
		switch errSubs {
		case sql.ErrNoRows, ErrAlreadyConfirmed:
			return subs, ErrDenyRetry
		default:
			return subs, ErrAllowRetry
		}
	}

	log.Infof("Found order %s", subs.OrderID)

	// STEP 2: query membership
	// For any errors, allow retry.
	member, errMember := tx.RetrieveMember(paywall.UserID{
		CompoundID: subs.CompoundID,
		FtcID:      subs.FtcID,
		UnionID:    subs.UnionID,
	})
	if errMember != nil {
		return subs, ErrAllowRetry
	}

	// STEP 3: validate the retrieved order.
	// This order might be invalid for upgrading.
	// OPTIONAL STEP: Mark the prorated orders.
	// For any errors, allow retry
	if subs.Usage == paywall.SubsKindUpgrade && member.Tier == enum.TierPremium && !member.IsExpired() {
		// In case of any error, just ignore it.
		err = tx.DuplicateUpgrade(subs.OrderID)

		if err != nil {
			_ = tx.rollback()
			return subs, paywall.ErrDuplicateUpgrading
		}

		if err := tx.commit(); err != nil {
			return subs, paywall.ErrDuplicateUpgrading
		}

		return subs, paywall.ErrDuplicateUpgrading
	}

	// STEP 4: Calculate order's confirmation time.
	// Populate the ConfirmedAt, StartDate and EndDate.
	// If there are calculation errors, allow retry.
	subs, err = subs.Confirm(member, confirmedAt)
	if err != nil {
		// Remember to rollback.
		_ = tx.tx.Rollback()
		return subs, ErrAllowRetry
	}

	log.Infof("Order confirmed: %s - %s", subs.StartDate, subs.EndDate)

	// STEP 5: Update confirmed order
	// For any errors, allow retry.
	updateErr := tx.ConfirmOrder(subs)
	if updateErr != nil {
		// Remember to rollback.
		_ = tx.tx.Rollback()
		return subs, ErrAllowRetry
	}

	// STEP 6: Build new membership from this order.
	// This error should allow retry.
	member, err = member.FromAliOrWx(subs)
	if err != nil {
		// Remember to rollback
		_ = tx.tx.Rollback()
		return subs, ErrAllowRetry
	}

	// STEP 7: Insert or update membership.
	// This error should allow retry
	if subs.Usage == paywall.SubsKindCreate {
		err := tx.CreateMember(member)
		if err != nil {
			return subs, ErrAllowRetry
		}
	} else {
		err := tx.UpdateMember(member)
		if err != nil {
			return subs, ErrAllowRetry
		}
	}

	//if err := tx.LinkUser(member); err != nil {
	//	_ = tx.rollback()
	//	return subs, err
	//}

	// Error here should allow retry.
	if err := tx.commit(); err != nil {
		log.Error(err)
		return subs, ErrAllowRetry
	}

	log.Info("ConfirmPayment finished.")

	return subs, nil
}
