package model

import (
	"database/sql"
	"time"

	"github.com/FTChinese/go-rest"
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
		s.UserID,
		s.FTCUserID,
		s.UnionID,
		s.ListPrice,
		s.NetPrice,
		s.TierToBuy,
		s.BillingCycle,
		s.PaymentMethod,
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
		&s.OrderID,
		&s.UserID,
		&s.FTCUserID,
		&s.UnionID,
		&s.TierToBuy,
		&s.BillingCycle,
		&s.ListPrice,
		&s.NetPrice,
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
		&subs.OrderID,
		&subs.UserID,
		&subs.FTCUserID,
		&subs.UnionID,
		&subs.TierToBuy,
		&subs.BillingCycle,
		&subs.ListPrice,
		&subs.NetPrice,
		&subs.PaymentMethod,
		&subs.CreatedAt,
		&subs.ConfirmedAt,
	)

	if errSubs != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)

		_ = tx.Rollback()

		if errSubs == sql.ErrNoRows {
			return subs, ErrOrderNotFound
		}
	}

	logger.Infof("Found order %s", subs.OrderID)

	// Already confirmed.
	if !subs.ConfirmedAt.IsZero() {
		logger.WithField("trace", "ConfirmPayment").Infof("Order %s is already confirmed", orderID)

		_ = tx.Rollback()
		// Already confirmed, do not retry any more.
		return subs, ErrAlreadyConfirmed
	}

	// A flag to determine insert a new member
	// or update an existing one.
	memberExists := true
	// Step 2: query membership expiration time to determine the order's start date, and then user start date to calculate end date.
	// The row is locked for update.
	var dur paywall.Duration
	errDur := env.db.QueryRow(
		env.stmtSelectExpireDate(),
		subs.UserID,
		subs.UnionID,
	).Scan(
		&dur.Timestamp,
		&dur.ExpireDate,
	)

	if errDur != nil {
		if errDur == sql.ErrNoRows {
			memberExists = false
		} else {
			logger.WithField("trace", "ConfirmPayment").Error(errDur)

			_ = tx.Rollback()
		}
	}

	logger.Infof("Membership for %s: %t", subs.UserID, memberExists)

	// For sql.ErrNoRows, the zero value of Duration
	// is still a valid value.
	subs, err = subs.ConfirmWithDuration(dur, confirmedAt)
	if err != nil {
		return subs, err
	}
	logger.Infof("Order confirmed: %s - %s", subs.StartDate, subs.EndDate)

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

	logger.Infof("Order updated")

	// Step 4: Insert a membership or update it in case of duplicate.
	if memberExists {
		logger.Infof("Extend an existing member")
		_, err := tx.Exec(env.stmtUpdateMember(),
			subs.FTCUserID,
			subs.UnionID,
			subs.TierToBuy,
			subs.BillingCycle,
			subs.EndDate,
			subs.UserID,
			subs.UnionID,
		)

		if err != nil {
			_ = tx.Rollback()

			logger.WithField("trace", "ConfirmPayment").Error(err)
		}
	} else {
		logger.Info("Create a new member")

		_, err := tx.Exec(env.stmtInsertMember(),
			subs.UserID,
			subs.UnionID,
			subs.FTCUserID,
			subs.UnionID,
			subs.TierToBuy,
			subs.BillingCycle,
			subs.EndDate,
		)
		if err != nil {
			_ = tx.Rollback()

			logger.WithField("trace", "ConfirmPayment").Error(err)
		}
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return subs, err
	}

	logger.Info("Confirm order finished")
	return subs, nil
}
