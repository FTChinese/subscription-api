package model

import (
	"database/sql"
	"time"

	gorest "github.com/FTChinese/go-rest"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

// IsSubsAllowed checks if this user is allowed to purchase a subscription.
// If a user is a valid member, and the membership is not expired, and not within the allowed renewal period, deny the request.
func (env Env) IsSubsAllowed(subs paywall.Subscription) (bool, error) {
	member, err := env.findMember(subs)

	if err != nil {
		// If this user is not a member yet.
		if err == sql.ErrNoRows {
			return true, nil
		}

		logger.WithField("trace", "IsSubsAllowed").Error(err)
		// If any other unknown error occurred
		return false, err
	}

	// Do not allow a subscribed user to change tiers.
	if subs.TierToBuy != member.Tier {
		logger.WithField("trace", "IsSubsAllowed").Error("Changing subscription tier is not supported.")
		return false, nil
	}

	// This user is/was a member.
	return member.CanRenew(subs.BillingCycle), nil
}

// SaveSubscription saves a new subscription order.
// At this moment, you should already know if this subscription is
// a renewal of a new one, based on current Membership's expire_date.
func (env Env) SaveSubscription(s paywall.Subscription, c gorest.ClientApp) error {

	_, err := env.db.Exec(
		env.stmtInsertSubs(),
		s.OrderID,
		s.ListPrice,
		s.NetPrice,
		s.UserID,
		s.LoginMethod,
		s.TierToBuy,
		s.BillingCycle,
		s.PaymentMethod,
		s.IsRenewal,
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
func (env Env) FindSubscription(orderID string) (paywall.Subscription, error) {

	var s paywall.Subscription
	err := env.db.QueryRow(
		env.stmtSelectSubs(),
		orderID,
	).Scan(
		&s.UserID,
		&s.OrderID,
		&s.ListPrice,
		&s.NetPrice,
		&s.LoginMethod,
		&s.TierToBuy,
		&s.BillingCycle,
		&s.PaymentMethod,
		&s.CreatedAt,
		&s.ConfirmedAt,
	)

	if err != nil {
		logger.WithField("trace", "FindSubscription").Error(err)
		return s, err
	}

	return s, nil
}

// ConfirmPayment handles payment notification with database locking.
// Returns the a complete Subscription to be used to compose an email.
// If returned error is ErrOrderNotFound or ErrAlreadyConfirmed, tell Wechat or Ali do not try any more; oterwise let them retry.
// Only when error is nil should be send a confirmation email.
func (env Env) ConfirmPayment(orderID string, confirmedAt time.Time) (paywall.Subscription, error) {

	var subs paywall.Subscription

	tx, err := env.db.Begin()
	if err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return subs, err
	}

	// Step 1: Find the subscription order by order id
	// The row is locked for update
	errSubs := env.db.QueryRow(
		env.stmtSelectSubsLock(),
		orderID,
	).Scan(
		&subs.UserID,
		&subs.OrderID,
		&subs.ListPrice,
		&subs.NetPrice,
		&subs.LoginMethod,
		&subs.TierToBuy,
		&subs.BillingCycle,
		&subs.PaymentMethod,
		&subs.CreatedAt,
		&subs.ConfirmedAt,
	)

	if errSubs != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)

		_ = tx.Rollback()
		// If this order does not exist, do not retry.
		if errSubs == sql.ErrNoRows {
			return subs, ErrOrderNotFound
		}
	}

	// Already confirmed.
	if !subs.ConfirmedAt.IsZero() {
		logger.WithField("trace", "ConfirmPayment").Infof("Order %s is already confirmed", orderID)

		_ = tx.Rollback()
		// Already confirmed, do not retry any more.
		return subs, ErrAlreadyConfirmed
	}

	logger.WithField("trace", "ConfirmPayment").Infof("Found order: %+v", subs)

	// Step 2: query membership expiration time to determine the order's start date, and then user start date to calculate end date.
	// The row is locked for update.
	var dur paywall.Duration
	errDur := env.db.QueryRow(
		env.stmtSelectExpLock(subs.IsWxLogin()),
		subs.UserID,
	).Scan(
		&dur.Timestamp,
		&dur.ExpireDate,
	)

	// If any db error occurred, tell API provider to retry.
	if errDur != nil && errDur != sql.ErrNoRows {
		logger.WithField("trace", "ConfirmPayment").Error(errDur)

		_ = tx.Rollback()
	}

	// For sql.ErrNoRows, `dur` is still a valid value.
	subs, err = subs.ConfirmWithDuration(dur, confirmedAt)
	if err != nil {
		return subs, err
	}

	logger.WithField("trace", "ConfirmPayment").Infof("Updated order: %+v", subs)

	// Step 3: Update subscription order with confirmation time, membership start date and end date.
	_, updateErr := tx.Exec(
		env.stmtUpdateSubs(),
		subs.IsRenewal,
		subs.ConfirmedAt,
		subs.StartDate,
		subs.EndDate,
		orderID,
	)

	// If any error occurred, retry.
	if updateErr != nil {
		_ = tx.Rollback()

		logger.WithField("trace", "ConfirmPayment").Error(updateErr)
	}

	// Step 4: Insert a membership or update it in case of duplicacy.
	_, createErr := tx.Exec(
		env.stmtInsertMember(),
		subs.UserID,
		subs.GetUnionID(),
		subs.TierToBuy,
		subs.BillingCycle,
		subs.EndDate,
		subs.TierToBuy,
		subs.BillingCycle,
		subs.EndDate,
	)

	if createErr != nil {
		_ = tx.Rollback()

		logger.WithField("trace", "ConfirmPayment").Error(err)
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return subs, err
	}

	return subs, nil
}
