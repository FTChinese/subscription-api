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
	CreatedAt     string // Only for retrieval
	ConfirmedAt   string
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

// SaveSubscription saves a new order
func (env Env) SaveSubscription(s Subscription, c util.RequestClient) error {
	query := `
	INSERT INTO premium.ftc_trade
	SET trade_no = ?,
		trade_price = ?,
		trade_amount = ?,
		tier_to_buy = ?,
		billing_cycle = ?,
		payment_method = ?,
		user_id = ?,
		client_type = ?,
		client_version = ?,
		created_utc = UTC_TIMESTAMP(),
		user_ip_bin = INET6_ATON(?)`

	_, err := env.DB.Exec(query,
		s.OrderID,
		s.Price,
		s.TotalAmount,
		string(s.TierToBuy),
		string(s.BillingCycle),
		string(s.PaymentMethod),
		s.UserID,
		c.ClientType,
		c.Version,
		c.UserIP,
	)

	if err != nil {
		logger.WithField("location", "New subscription").Error(err)
		return err
	}

	return nil
}

// FindSubscription tries to find an order
func (env Env) FindSubscription(orderID string) (Subscription, error) {
	query := `
	SELECT trade_no AS orderId,
		trade_price AS price,
		trade_amount AS totalAmount,
		user_id AS userId,
		IFNULL(tier_to_buy, '') AS tierToBuy,
		IFNULL(billing_cycle, '') AS billingCycle,
		IFNULL(payment_method, '') AS paymentMethod,
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
		&s.CreatedAt,
		&s.ConfirmedAt,
	)

	if err != nil {
		logger.WithField("location", "Find subscription").Error(err)
		return s, err
	}

	return s, nil
}

// ConfirmSubscription marks an order as completed and create a member.
// Confirm order and create/renew a new member should be an all-or-nothing operation.
// Or update membership duration.
// NOTE: The passed in Subscription must be one retrieved from database. Otherwise you should never call this method.
func (env Env) ConfirmSubscription(s Subscription, confirmTime time.Time) error {
	// Subscription confirmation time.
	confirmedAt := util.SQLDatetimeUTC.FromTime(confirmTime)

	// Build membership based on subscription.
	m := env.buildMembership(s, confirmTime)

	tx, err := env.DB.Begin()
	if err != nil {
		logger.WithField("location", "Confirm subscripiton").Error(err)
		return err
	}
	stmtUpdate := `
	UPDATE premium.ftc_trade
	SET confirmed_utc = ?
	WHERE trade_no = ?
	LIMIT 1`

	_, updateErr := tx.Exec(stmtUpdate,
		confirmedAt,
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
		m.UserID,
		string(m.Tier),
		string(m.Cycle),
		m.Expire,
		string(m.Tier),
		string(m.Cycle),
		m.Expire,
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
