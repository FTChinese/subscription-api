package model

import (
	"strconv"
	"time"

	"gitlab.com/ftchinese/subscription-api/util"
)

// Subscription contains the details of a user's action to place an order.
type Subscription struct {
	OrderID       string
	TierToBuy     MemberTier
	BillingCycle  BillingCycle
	Price         float64
	TotalAmount   float64
	PaymentMethod PaymentMethod
	Currency      string
	CreatedAt     string // When the order is created.
	ConfirmedAt   string // When the payment is confirmed.
	IsRenewal     bool   // If this order is used to renew membership
	StartDate     string // Membership start date for this order
	EndDate       string // Membership end date for this order
	UserID        string
}

// WxTotalFee converts TotalAmount to int64 in cent for comparison with wx notification.
func (s Subscription) WxTotalFee() int64 {
	return int64(s.TotalAmount * 100)
}

// AliTotalAmount converts TotalAmount to ailpay format
func (s Subscription) AliTotalAmount() string {
	return strconv.FormatFloat(s.TotalAmount, 'f', 2, 32)
}

// DeduceExpireTime deduces membership expiration time based on when it is confirmed and the billing cycle.
// Add one day more to accomodate timezone change
func (s Subscription) DeduceExpireTime(t time.Time) time.Time {
	switch s.BillingCycle {
	case Yearly:
		return t.AddDate(1, 0, 1)

	case Monthly:
		return t.AddDate(0, 1, 1)
	}

	return t
}

// RenewExpireDate extends member's expiration date depending on subscription
// func (s Subscription) RenewExpireDate(previous string) string {
// 	expire, err := util.ParseSQLDate(previous)
// 	if err != nil {
// 		return previous
// 	}

// 	switch s.BillingCycle {
// 	case Yearly:
// 		expire = expire.AddDate(1, 0, 1)

// 	case Monthly:
// 		expire = expire.AddDate(0, 1, 1)
// 	}

// 	return util.SQLDateUTC.FromTime(expire)
// }

// SaveSubscription saves a new subscription order.
// At this moment, you should already know if this subscription is
// a renewal of a new one, based on current Membership's expire_date.
func (env Env) SaveSubscription(s Subscription, c util.RequestClient) error {
	query := `
	INSERT INTO premium.ftc_trade
	SET trade_no = ?,
		trade_price = ?,
		trade_amount = ?,
		user_id = ?,
		tier_to_buy = ?,
		billing_cycle = ?,
		payment_method = ?,
		is_renewal = ?,
		created_utc = UTC_TIMESTAMP(),
		client_type = ?,
		client_version = ?,
		user_ip_bin = INET6_ATON(?),
		user_agent = NULLIF(?, '')`

	_, err := env.DB.Exec(query,
		s.OrderID,
		s.Price,
		s.TotalAmount,
		s.UserID,
		string(s.TierToBuy),
		string(s.BillingCycle),
		string(s.PaymentMethod),
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

// FindSubscription tries to find an order to verify the authenticity of a subscription order.
func (env Env) FindSubscription(orderID string) (Subscription, error) {
	query := `
	SELECT trade_no AS orderId,
		trade_price AS price,
		trade_amount AS totalAmount,
		user_id AS userId,
		IFNULL(tier_to_buy, '') AS tierToBuy,
		IFNULL(billing_cycle, '') AS billingCycle,
		IFNULL(payment_method, '') AS paymentMethod,
		is_renewal AS isRenewal,
		created_utc AS createdAt,
		IFNULL(confirmed_utc, '') AS confirmedAt
	FROM premium.ftc_trade
	WHERE trade_no = ?
	LIMIT 1`

	var s Subscription
	err := env.DB.QueryRow(query, orderID).Scan(
		&s.OrderID,
		&s.Price,
		&s.TotalAmount,
		&s.UserID,
		&s.TierToBuy,
		&s.BillingCycle,
		&s.PaymentMethod,
		&s.IsRenewal,
		&s.CreatedAt,
		&s.ConfirmedAt,
	)

	if err != nil {
		logger.WithField("location", "Find subscription").Error(err)
		return s, err
	}

	return s, nil
}

// ConfirmSubscription marks an order as completed and create a member or renew membership.
// Confirm order and create/renew a new member should be an all-or-nothing operation.
// Or update membership duration.
// NOTE: The passed in Subscription must be one retrieved from database. Otherwise you should never call this method.
func (env Env) ConfirmSubscription(s Subscription, confirmTime time.Time) error {
	// SQL DATETIME string for confirmation time.
	confirmedUTC := util.SQLDatetimeUTC.FromTime(confirmTime)

	// By default we assume this order will take effect from its
	// confirmation time.
	startTime := confirmTime

	// If this order is used for renewal, copy current memership's
	// expire_date directly to s.StartDate.
	if s.IsRenewal {
		member, err := env.FindMember(s.UserID)

		// If membership is found. Use its Expire date as startTime
		if err == nil {
			expireTime, err := util.ParseSQLDate(member.Expire)
			// If Expire date could be properly parsed.
			if err != nil {
				startTime = expireTime
			}
		}
	}

	expireTime := s.DeduceExpireTime(startTime)

	// SQL DATE string for start_date and end_date
	startDate := util.SQLDateUTC.FromTime(startTime)
	endDate := util.SQLDateUTC.FromTime(expireTime)

	// Build membership based on subscription.
	// m := env.buildMembership(s, confirmTime)

	tx, err := env.DB.Begin()
	if err != nil {
		logger.WithField("location", "Confirm subscripiton").Error(err)
		return err
	}
	stmtUpdate := `
	UPDATE premium.ftc_trade
	SET confirmed_utc = ?,
		start_date = ?,
		end_date = ?
	WHERE trade_no = ?
	LIMIT 1`

	_, updateErr := tx.Exec(stmtUpdate,
		confirmedUTC,
		startDate,
		endDate,
		s.OrderID,
	)

	if updateErr != nil {
		_ = tx.Rollback()
		logger.WithField("location", "Update subscription confirmation time").Error(err)
	}

	stmtCreate := `
	INSERT INTO premium.ftc_vip
	SET vip_id = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?
	ON DUPLICATE KEY UPDATE
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?`

	_, createErr := tx.Exec(stmtCreate,
		s.UserID,
		string(s.TierToBuy),
		string(s.BillingCycle),
		endDate,
		string(s.TierToBuy),
		string(s.BillingCycle),
		endDate,
	)

	if createErr != nil {
		_ = tx.Rollback()

		logger.WithField("location", "Create or renew membership").Error(err)
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("location", "Commit transaction to commit and create/renew membership").Error(err)
		return err
	}

	return nil
}
