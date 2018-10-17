package model

import "gitlab.com/ftchinese/subscription-api/util"

// SubscribeOrder records the details of an order a user placed.
type SubscribeOrder struct {
	OrderID       string
	TierToBuy     util.MemberTier
	BillingCycle  util.BillingCycle
	Price         int64
	TotalAmount   int64
	PaymentMethod util.PaymentMethod
	Currency      string
	CreatedAt     string
	ConfirmedAt   string
	UserID        string
}

// NewOrder saves a new order
func (env Env) NewOrder(order SubscribeOrder, c util.RequestClient) error {
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
		order.OrderID,
		order.Price,
		order.TotalAmount,
		string(order.TierToBuy),
		string(order.BillingCycle),
		string(order.PaymentMethod),
		order.UserID,
		c.ClientType,
		c.Version,
		order.CreatedAt,
		c.UserIP,
	)

	if err != nil {
		return err
	}

	return nil
}
