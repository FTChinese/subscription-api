package model

import (
	"database/sql"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/util"
	"time"

	"gitlab.com/ftchinese/subscription-api/paywall"
)

// SaveSubscription saves a new subscription order.
func (env Env) SaveSubscription(s paywall.Subscription, c util.ClientApp) error {

	orderIDs := s.UpgradeSourceIDs()

	_, err := env.db.Exec(
		env.query.InsertSubs(),
		s.OrderID,
		s.CompoundID,
		s.FtcID,
		s.UnionID,
		s.ListPrice,
		s.NetPrice,
		s.TierToBuy,
		s.BillingCycle,
		s.CycleCount,
		s.ExtraDays,
		s.Kind,
		null.NewString(orderIDs, orderIDs != ""),
		s.UpgradeBalance,
		s.PaymentMethod,
		s.WxAppID,
		c.ClientType,
		c.Version,
		c.UserIP,
		c.UserAgent)

	if err != nil {
		logger.WithField("trace", "SaveSubscription").Error(err)
		return err
	}

	return nil
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
		logger.WithField("trace", "FindSubsCharge").Error(err)
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

	tx, err := env.BeginMemberTx()
	if err != nil {
		logger.WithField("trace", "Env.ConfirmPayment").Error(err)
		return paywall.Subscription{}, ErrAllowRetry
	}

	// Step 1: Find the subscription order by order id
	// The row is locked for update.
	// If the order is not found, or is already confirmed,
	// tell provider not sending notification any longer;
	// otherwise, allow retry.
	subs, errSubs := tx.RetrieveOrder(orderID)
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

	// STEP 2: query membership
	// For any errors, allow retry.
	member, errMember := tx.RetrieveMember(subs)
	if errMember != nil {
		return subs, ErrAllowRetry
	}

	// STEP 3: validate the retrieved order.
	// This order might be invalid for upgrading.
	errInvalid := subs.Validate(member)
	// If the order is invalid, record the reason and
	// stop any further processing.
	if errInvalid != nil {
		_ = tx.InvalidUpgrade(subs.OrderID, errInvalid)
		_ = tx.rollback()
		return subs, ErrDenyRetry
	}

	// STEP 4: Calculate order's confirmation time.
	// Populate the ConfirmedAt, StartDate and EndDate.
	// If there are calculation errors, allow retry.
	subs, err = subs.ConfirmWithMember(member, confirmedAt)
	if err != nil {
		// Remember to rollback.
		_ = tx.tx.Rollback()
		return subs, ErrAllowRetry
	}

	logger.
		WithField("trace", "Env.ConfirmPayment").
		Infof("Order confirmed: %s - %s", subs.StartDate, subs.EndDate)

	// STEP 5: Update confirmed order
	// For any errors, allow retry.
	updateErr := tx.ConfirmOrder(subs)
	if updateErr != nil {
		// Remember to rollback.
		_ = tx.tx.Rollback()
		return subs, ErrAllowRetry
	}

	// OPTIONAL STEP: Mark the prorated orders.
	// For any errors, allow retry
	if subs.Kind == paywall.SubsKindUpgrade {
		updateErr := tx.MarkOrdersProrated(subs)
		if updateErr != nil {
			return subs, ErrAllowRetry
		}
	}

	// STEP 6: Build new membership from this order.
	// This error should allow retry.
	member, err = subs.BuildMembership()
	if err != nil {
		// Remember to rollback
		_ = tx.tx.Rollback()
		return subs, ErrAllowRetry
	}

	// STEP 7: Insert or update membership.
	// This error should allow retry
	upsertErr := tx.UpsertMember(member)
	if upsertErr != nil {
		return subs, ErrAllowRetry
	}

	// Error here should allow retry.
	if err := tx.commit(); err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return subs, ErrAllowRetry
	}

	logger.Info("Confirm payment finished")
	return subs, nil
}

// FindProration loads all orders that are in active user or
// not consumed yet and calculate the unused portion of
// each order.
//func (env Env) FindProration(u paywall.User) ([]paywall.Proration, error) {
//
//	rows, err := env.db.Query(
//		env.query.ProratedOrders(),
//		u.CompoundID,
//		u.UnionID)
//	if err != nil {
//		logger.WithField("trace", "FindProration").Error(err)
//		return nil, err
//	}
//	defer rows.Close()
//
//	orders := make([]paywall.Proration, 0)
//	for rows.Next() {
//		var o paywall.Proration
//
//		err := rows.Scan(
//			&o.OrderID,
//			&o.Balance,
//			&o.StartDate,
//			&o.EndDate)
//
//		if err != nil {
//			logger.WithField("trace", "FindProration").Error(err)
//			continue
//		}
//
//		orders = append(orders, o)
//	}
//
//	if err := rows.Err(); err != nil {
//		logger.WithField("trace", "FindProration").Error(err)
//		return nil, err
//	}
//
//	return orders, nil
//}

// FindUnusedOrders retrieves all orders that has unused portions.
func (env Env) FindUnusedOrders(u paywall.User) ([]paywall.UnusedOrder, error) {
	rows, err := env.db.Query(
		env.query.UnusedOrders(),
		u.CompoundID,
		u.UnionID)
	if err != nil {
		logger.WithField("trace", "FindUnusedOrders").Error(err)
		return nil, err
	}
	defer rows.Close()

	orders := make([]paywall.UnusedOrder, 0)
	for rows.Next() {
		var o paywall.UnusedOrder

		err := rows.Scan(
			&o.ID,
			&o.NetPrice,
			&o.StartDate,
			&o.EndDate)

		if err != nil {
			logger.WithField("trace", "FindUnusedOrders").Error(err)
			return nil, err
		}

		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		logger.WithField("trace", "FindUnusedOrders").Error(err)
		return nil, err
	}

	return orders, nil
}

// BuildUpgradePlan creates upgrade plan based on user's
// previous orders.
// See [Env.GetCurrentPlan]
func (env Env) BuildUpgradePlan(u paywall.User, p paywall.Plan) (paywall.UpgradePlan, error) {
	if env.sandbox {
		return paywall.GetSandboxUpgrade(), nil
	}

	//orders, err := env.FindProration(u)
	orders, err := env.FindUnusedOrders(u)
	if err != nil {
		return paywall.UpgradePlan{}, err
	}

	up := paywall.NewUpgradePlan(p).
		//SetProration(orders).
		SetBalance(orders).
		CalculatePayable()

	return up, nil
}
