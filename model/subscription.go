package model

import (
	"time"

	"gitlab.com/ftchinese/subscription-api/util"
)

// Subscription contains the details of a user's action to place an order.
type Subscription struct {
	OrderID       string
	TierToBuy     MemberTier
	BillingCycle  BillingCycle
	Price         int64
	TotalAmount   int64
	PaymentMethod PaymentMethod
	Currency      string
	CreatedAt     string
	ConfirmedAt   string
	UserID        string
}

// DeduceExpireTime deduces membership expiration time based on when it is confirmed and the billing cycle.
func (o Subscription) DeduceExpireTime(t time.Time) time.Time {
	switch o.BillingCycle {
	case Yearly:
		return t.AddDate(1, 0, 0)

	case Monthly:
		return t.AddDate(0, 1, 0)
	}

	return t
}

// NewOrder saves a new order
func (env Env) NewOrder(s Subscription, c util.RequestClient) error {
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
		s.CreatedAt,
		c.UserIP,
	)

	if err != nil {
		return err
	}

	return nil
}

// RetrieveOrder tries to find an order
func (env Env) RetrieveOrder(orderID string) (Subscription, error) {
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
		s.OrderID,
		s.Price,
		s.TotalAmount,
		s.UserID,
		s.TierToBuy,
		s.BillingCycle,
		s.PaymentMethod,
		s.CreatedAt,
		s.ConfirmedAt,
	)

	if err != nil {
		return s, err
	}

	return s, nil
}

// ConfirmOrder marks an order as completed and create a member.
// Confirm order and create/renew a new member should be an all-or-nothing operation.
// Or update membership duration.
func (env Env) ConfirmOrder(order Subscription, confirmTime time.Time) error {

	confirmedAt := util.SQLDatetimeUTC.FromTime(confirmTime)

	m := env.NewMemberFromOrder(order, confirmTime)

	tx, err := env.DB.Begin()
	if err != nil {
		return err
	}
	stmtUpdate := `
	UPDATE premium.ftc_trade
	SET confirmed_utc = ?
	WHERE trade_no = ?
	LIMIT 1`

	_, updateErr := tx.Exec(stmtUpdate,
		confirmedAt,
		order.OrderID,
	)

	if updateErr != nil {
		_ = tx.Rollback()
	}

	stmtCreate := `
	INSERT INTO premium.ftc_vip
	SET vip_id = ?,
		member_tier = ?,
		billing_cycle = ?,
		start_utc = ?,
		expire_utc = ?
	ON DUPLICATE KEY UPDATE
		member_tier = ?,
		billing_cycle = ?,
		start_utc = ?,
		expire_utc = ?`

	_, createErr := tx.Exec(stmtCreate,
		m.UserID,
		string(m.Tier),
		string(m.Cycle),
		m.Start,
		m.Expire,
		string(m.Tier),
		string(m.Cycle),
		m.Start,
		m.Expire,
	)

	if createErr != nil {
		_ = tx.Rollback()
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
