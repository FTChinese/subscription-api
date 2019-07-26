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
	user paywall.AccountID,
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

	// Use the found member to determine order category.
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

		err = otx.SaveUpgrade(subs.ID, upgrade)
		if err != nil {
			_ = otx.rollback()
			return subs, err
		}

		log.Infof("Saved upgrade scheme %s", upgrade.ID)

		err = otx.SetLastUpgradeID(upgrade)
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
		&c.Amount,
		&c.IsConfirmed,
	)

	if err != nil {
		logger.WithField("trace", "Env.FindSubsCharge").Error(err)
		return c, err
	}

	c.NetPrice = c.Amount
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
func (env Env) ConfirmPayment(orderID string, confirmedAt time.Time) (paywall.Subscription, *paywall.ConfirmationResult) {

	log := logger.WithField("trace", "Env.ConfirmPayment")

	tx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.Subscription{}, paywall.NewConfirmationFailed(orderID, err, true)
	}

	// Step 1: Find the subscription order by order id
	// The row is locked for update.
	// If the order is not found, or is already confirmed,
	// tell provider not sending notification any longer;
	// otherwise, allow retry.
	subs, err := tx.RetrieveOrder(orderID)
	if err != nil {
		log.Error(err)

		_ = tx.rollback()

		return subs, paywall.NewConfirmationFailed(orderID, err, err != sql.ErrNoRows)
	}

	if subs.IsConfirmed {
		_ = tx.rollback()
		return subs, paywall.NewConfirmationFailed(orderID, util.ErrAlreadyConfirmed, false)
	}

	log.Infof("Found order %s", subs.ID)

	// STEP 2: query membership
	// For any errors, allow retry.
	member, err := tx.RetrieveMember(subs.User)
	if err != nil {
		log.Error(err)
		_ = tx.rollback()
		return subs, paywall.NewConfirmationFailed(orderID, err, true)
	}

	// STEP 3: validate the retrieved order.
	// This order might be invalid for upgrading.
	// If user is already a premium member and this order is used
	// for upgrading, decline retry.
	if subs.Usage == paywall.SubsKindUpgrade && member.IsValidPremium() {
		_ = tx.rollback()
		return subs, paywall.NewConfirmationFailed(orderID, util.ErrDuplicateUpgrading, false)
	}

	// STEP 4: Calculate order's confirmation time.
	// Populate the ConfirmedAt, StartDate and EndDate.
	// If there are calculation errors, allow retry.
	subs, err = subs.Confirm(member, confirmedAt)
	if err != nil {
		log.Error(err)
		_ = tx.rollback()
		return subs, paywall.NewConfirmationFailed(orderID, err, true)
	}

	log.Infof("Order confirmed: %s - %s", subs.StartDate, subs.EndDate)

	// STEP 5: Update confirmed order
	// For any errors, allow retry.
	err = tx.ConfirmOrder(subs)
	if err != nil {
		log.Error(err)
		// Remember to rollback.
		_ = tx.rollback()
		return subs, paywall.NewConfirmationFailed(orderID, err, true)
	}

	// STEP 6: Build new membership from this order.
	// This error should allow retry.
	member, err = member.FromAliOrWx(subs)
	if err != nil {
		log.Error(err)
		_ = tx.rollback()
		return subs, paywall.NewConfirmationFailed(orderID, err, true)
	}

	// STEP 7: Insert or update membership.
	// This error should allow retry
	if subs.Usage == paywall.SubsKindCreate {
		err := tx.CreateMember(member)
		if err != nil {
			log.Error(err)
			_ = tx.rollback()
			return subs, paywall.NewConfirmationFailed(orderID, err, true)
		}
	} else {
		err := tx.UpdateMember(member)
		if err != nil {
			log.Error(err)
			_ = tx.rollback()
			return subs, paywall.NewConfirmationFailed(orderID, err, true)
		}
	}

	// Error here should allow retry.
	if err := tx.commit(); err != nil {
		log.Error(err)
		return subs, paywall.NewConfirmationFailed(orderID, err, true)
	}

	log.Info("ConfirmPayment finished.")

	return subs, nil
}

func (env Env) SaveConfirmationResult(r *paywall.ConfirmationResult) error {
	_, err := env.db.Exec(env.query.ConfirmationResult(),
		r.OrderID,
		r.Succeeded,
		r.Failed)

	if err != nil {
		return err
	}

	return nil
}
