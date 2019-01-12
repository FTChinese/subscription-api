package model

import (
	"database/sql"
	"time"

	"github.com/objcoding/wxpay"

	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
)

// IsSubsAllowed checks if this user is allowed to purchase a subscritpion.
// If a user is a valid member, and the membership is not expired, and not within the allowed renewal period, deny the request.
func (env Env) IsSubsAllowed(subs paywall.Subscription) (bool, error) {
	member, err := env.findMember(subs)

	if err != nil {
		// If this user is not a member yet.
		if err == sql.ErrNoRows {
			return true, nil
		}

		logger.WithField("trace", "IsSubsAllowed").Error(err)
		// If any other unkonw error occurred
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
func (env Env) SaveSubscription(s paywall.Subscription, c util.ClientApp) error {
	query := `
	INSERT INTO premium.ftc_trade
	SET trade_no = ?,
		trade_price = ?,
		trade_amount = ?,
		user_id = ?,
		login_method = ?,
		tier_to_buy = ?,
		billing_cycle = ?,
		payment_method = ?,
		is_renewal = ?,
		created_utc = UTC_TIMESTAMP(),
		client_type = ?,
		client_version = ?,
		user_ip_bin = INET6_ATON(?),
		user_agent = ?`

	_, err := env.DB.Exec(query,
		s.OrderID,
		s.Price,
		s.TotalAmount,
		s.UserID,
		s.LoginMethod,
		s.TierToBuy,
		s.BillingCycle,
		s.PaymentMethod,
		s.IsRenewal,
		c.ClientType,
		c.Version,
		c.UserIP,
		c.UserAgent,
	)

	if err != nil {
		logger.WithField("location", "New subscription").Error(err)
		return err
	}

	return nil
}

// VerifyWxNotification checks if price match, if already confirmed.
func (env Env) VerifyWxNotification(p wxpay.Params) error {
	orderID := p.GetString("out_trade_no")
	totalFee := p.GetInt64("total_fee")

	query := `
	SELECT trade_amount AS totalAmount
		confirmed_utc AS confirmedAt
	FROM premium.ftc_trade
	WHERE trade_no = ?
	LIMIT 1`

	var amount float64
	var confirmedAt util.Time
	err := env.DB.QueryRow(query, orderID).Scan(
		&amount,
		&confirmedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderNotFound
		}
		return err
	}

	if !confirmedAt.IsZero() {
		logger.WithField("trace", "VerifyWxNotification").Error(ErrAlreadyConfirmed)

		return ErrAlreadyConfirmed
	}

	price := int64(amount * 100)

	if price != totalFee {
		logger.WithField("trace", "VerifyWxNotification").Infof("Paid price does not match. Should be %d, actual %d", price, totalFee)

		return ErrPriceMismatch
	}

	return nil
}

// FindSubscription tries to find an order to verify the authenticity of a subscription order.
func (env Env) FindSubscription(orderID string) (paywall.Subscription, error) {
	query := `
	SELECT trade_no AS orderId,
		trade_price AS price,
		trade_amount AS totalAmount,
		user_id AS userId,
		login_method AS loginMethod,
		tier_to_buy AS tierToBuy,
		billing_cycle AS billingCycle,
		payment_method AS paymentMethod,
		created_utc AS createdAt,
		confirmed_utc AS confirmedAt,
		start_date AS startDate,
		end_date AS endDate
	FROM premium.ftc_trade
	WHERE trade_no = ?
	LIMIT 1`

	var s paywall.Subscription
	err := env.DB.QueryRow(query, orderID).Scan(
		&s.OrderID,
		&s.Price,
		&s.TotalAmount,
		&s.UserID,
		&s.LoginMethod,
		&s.TierToBuy,
		&s.BillingCycle,
		&s.PaymentMethod,
		&s.CreatedAt,
		&s.ConfirmedAt,
		&s.StartDate,
		&s.EndDate,
	)

	if err != nil {
		logger.WithField("trace", "FindSubscription").Error(err)
		return s, err
	}

	return s, nil
}

// ConfirmPayment handles payment notification with database locking.
func (env Env) ConfirmPayment(orderID string, confirmedAt time.Time) (paywall.Subscription, error) {

	var subs paywall.Subscription
	var startTime time.Time

	tx, err := env.DB.Begin()
	if err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return subs, err
	}

	errSubs := env.DB.QueryRow(stmtSubsLock, orderID).Scan(
		&subs.UserID,
		&subs.OrderID,
		&subs.LoginMethod,
		&subs.TierToBuy,
		&subs.BillingCycle,
		&subs.Price,
		&subs.TotalAmount,
		&subs.CreatedAt,
		&subs.ConfirmedAt,
	)

	if errSubs != nil {
		_ = tx.Rollback()
		if errSubs == sql.ErrNoRows {
			return subs, ErrOrderNotFound
		}
		return subs, errSubs
	}

	// Already confirmed.
	if !subs.ConfirmedAt.IsZero() {
		logger.WithField("trace", "ConfirmPayment").Infof("Order %s is already confirmed", orderID)

		_ = tx.Rollback()
		return subs, ErrAlreadyConfirmed
	}

	logger.WithField("trace", "ConfirmPayment").Infof("Found order: %+v", subs)

	// Add confirmation time.
	subs.ConfirmedAt = util.TimeFrom(confirmedAt)

	// Start query membership expiration time.
	queryDuration := subs.StmtMemberDuration()

	var dur paywall.Duration
	errDur := env.DB.QueryRow(queryDuration, orderID).Scan(
		&dur.Timestamp,
		&dur.ExpireDate,
	)

	if errDur != nil {
		// If no current membership is found for this order, confirmation time is the membership's start time.
		if errDur == sql.ErrNoRows {
			logger.WithField("trace", "ConfirmPayment").Infof("Member duration for user %s is not found", subs.UserID)

			subs.IsRenewal = false
			startTime = confirmedAt
		} else {
			_ = tx.Rollback()
			return subs, err
		}
	}

	dur.NormalizeDate()
	// If membership is found, test if it is expired.
	if dur.IsExpired() {
		subs.IsRenewal = false
		startTime = confirmedAt
	} else {
		subs.IsRenewal = true
		startTime = dur.ExpireDate.Time
	}

	subs, err = subs.WithStartTime(startTime)
	if err != nil {
		return subs, err
	}

	logger.WithField("trace", "ConfirmPayment").Infof("Updated order: %+v", subs)

	// Update subscription order.
	_, updateErr := tx.Exec(stmtUpdateSubs,
		subs.IsRenewal,
		subs.ConfirmedAt,
		subs.StartDate,
		subs.EndDate,
		orderID,
	)

	if updateErr != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "ConfirmPayment").Error(err)
	}

	// Create or extend membership.
	_, createErr := tx.Exec(stmtCreateMember,
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
