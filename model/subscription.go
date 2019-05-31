package model

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/util"
	"time"

	"gitlab.com/ftchinese/subscription-api/paywall"
)

// CreateOrder creates an order for a user with the selected plan and payment method.
func (env Env) CreateOrder(user paywall.User, plan paywall.Plan, payMethod enum.PayMethod, clientApp util.ClientApp, wxAppId null.String) (paywall.Subscription, error) {
	otx, err := env.BeginOrderTx()
	if err != nil {
		logger.WithField("trace", "Env.CreateOrder").Error(err)
		return paywall.Subscription{}, err
	}

	// Find the member
	member, err := otx.RetrieveMember(user)
	if err != nil && err != sql.ErrNoRows {
		_ = otx.rollback()
		return paywall.Subscription{}, err
	}

	// User the found member to determine order category.
	// Ignore Golang warning here. It is safe here to user
	// the zero value of membership since it is not a pointer.
	subsKind, err := member.SubsKind(plan)
	if err != nil {
		return paywall.Subscription{}, err
	}

	// Create order
	var subs paywall.Subscription
	if subsKind == paywall.SubsKindUpgrade {
		subs, err = otx.BuildUpgradeOrder(user, plan)
		if err != nil {
			_ = otx.rollback()
			return subs, err
		}
	} else {
		subs, err = paywall.NewSubs(user, plan)
		if err != nil {
			return subs, err
		}
		subs.Kind = subsKind
	}
	subs.PaymentMethod = payMethod
	subs.WxAppID = wxAppId

	// Save Order
	if env.sandbox {
		subs.NetPrice = 0.01
	}

	err = otx.SaveOrder(subs, clientApp)
	if err != nil {
		_ = otx.rollback()
		return subs, err
	}

	// Save upgrade source order.
	if subsKind == paywall.SubsKindUpgrade {
		err := otx.SaveUpgradeSource(subs.OrderID, subs.UpgradeSource)
		if err != nil {
			_ = otx.rollback()
			return subs, err
		}
	}

	if err := otx.commit(); err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return paywall.Subscription{}, err
	}

	// For sandbox.
	if env.sandbox {
		subs.NetPrice = 0.01
		return subs, nil
	}
	// Return the order
	return subs, nil
}

// UpgradePlan builds upgrade preview for a standard user who
// is trying to upgrade to premium.
// DO remember to rollback!
func (env Env) UpgradePlan(user paywall.User) (paywall.UpgradePlan, error) {
	otx, err := env.BeginOrderTx()
	if err != nil {
		logger.WithField("trace", "Env.CreateOrder").Error(err)
		return paywall.UpgradePlan{}, err
	}

	member, err := otx.RetrieveMember(user)
	// If membership is not found for this user, deny upgrading.
	if err != nil {
		_ = otx.rollback()
		return paywall.UpgradePlan{}, err
	}

	if member.Tier == enum.TierPremium {
		_ = otx.rollback()
		return paywall.UpgradePlan{}, ErrAlreadyUpgraded
	}

	orders, err := otx.FindUnusedOrders(user)
	if err != nil {
		_ = otx.rollback()
		return paywall.UpgradePlan{}, err
	}

	plan, _ := env.GetCurrentPricing().FindPlan(enum.TierPremium.String(), enum.CycleYear.String())

	up := paywall.NewUpgradePlan(plan).
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
func (env Env) DirectUpgradeOrder(user paywall.User, plan paywall.UpgradePlan, clientApp util.ClientApp) (paywall.Subscription, error) {

	subs, err := paywall.NewSubsUpgrade(user, plan)
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

	err = otx.SaveOrder(subs, clientApp)
	if err != nil {
		_ = otx.rollback()
		return subs, err
	}

	err = otx.SaveUpgradeSource(subs.OrderID, subs.UpgradeSource)
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

	mtx, err := env.BeginMemberTx()
	if err != nil {
		logger.WithField("trace", "Env.ConfirmPayment").Error(err)
		return paywall.Subscription{}, ErrAllowRetry
	}

	// Step 1: Find the subscription order by order id
	// The row is locked for update.
	// If the order is not found, or is already confirmed,
	// tell provider not sending notification any longer;
	// otherwise, allow retry.
	subs, errSubs := mtx.RetrieveOrder(orderID)
	if errSubs != nil {
		switch errSubs {
		case sql.ErrNoRows, ErrAlreadyConfirmed:
			return subs, ErrDenyRetry
		default:
			return subs, ErrAllowRetry
		}
	}

	logger.
		WithField("trace", "Env.ConfirmPayment").
		Infof("Found order %s", subs.OrderID)

	// If this is a upgrade order, find out the orders prorated
	// to the account balance.
	if subs.Kind == paywall.SubsKindUpgrade {
		//  nil value is still valid.
		srcIDs, err := mtx.RetrieveUpgradeSource(subs.OrderID)
		if err == nil {
			subs.UpgradeSource = srcIDs
		}
	}

	// STEP 2: query membership
	// For any errors, allow retry.
	member, errMember := mtx.RetrieveMember(subs)
	if errMember != nil {
		return subs, ErrAllowRetry
	}

	// STEP 3: validate the retrieved order.
	// This order might be invalid for upgrading.
	errInvalid := subs.Validate(member)
	// If the order is invalid, record the reason and
	// stop any further processing.
	if errInvalid != nil {
		_ = mtx.InvalidUpgrade(subs.OrderID, errInvalid)
		_ = mtx.rollback()
		return subs, ErrDenyRetry
	}

	// STEP 4: Calculate order's confirmation time.
	// Populate the ConfirmedAt, StartDate and EndDate.
	// If there are calculation errors, allow retry.
	subs, err = subs.ConfirmWithMember(member, confirmedAt)
	if err != nil {
		// Remember to rollback.
		_ = mtx.tx.Rollback()
		return subs, ErrAllowRetry
	}

	logger.
		WithField("trace", "Env.ConfirmPayment").
		Infof("Order confirmed: %s - %s", subs.StartDate, subs.EndDate)

	// STEP 5: Update confirmed order
	// For any errors, allow retry.
	updateErr := mtx.ConfirmOrder(subs)
	if updateErr != nil {
		// Remember to rollback.
		_ = mtx.tx.Rollback()
		return subs, ErrAllowRetry
	}

	// OPTIONAL STEP: Mark the prorated orders.
	// For any errors, allow retry
	if subs.Kind == paywall.SubsKindUpgrade {
		// In case of any error, just ignore it.
		_ = mtx.ConfirmUpgradeSource(subs.OrderID)
	}

	// STEP 6: Build new membership from this order.
	// This error should allow retry.
	member, err = subs.BuildMembership()
	if err != nil {
		// Remember to rollback
		_ = mtx.tx.Rollback()
		return subs, ErrAllowRetry
	}

	// STEP 7: Insert or update membership.
	// This error should allow retry
	upsertErr := mtx.UpsertMember(member)
	if upsertErr != nil {
		return subs, ErrAllowRetry
	}

	// Error here should allow retry.
	if err := mtx.commit(); err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return subs, ErrAllowRetry
	}

	logger.Info("Confirm payment finished")
	return subs, nil
}
