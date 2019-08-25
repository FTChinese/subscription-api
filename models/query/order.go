package query

import "fmt"

func (b Builder) InsertClientApp() string {
	return fmt.Sprintf(`
	INSERT INTO %s.client
	SET order_id = :order_id,
		client_type = :client_type,
		client_version = :client_version,
		user_ip_bin = INET6_ATON(:user_ip),
		user_agent = :user_agent`, b.MemberDB())
}

// Statement to insert a subscription order.
func (b Builder) InsertOrder() string {
	return fmt.Sprintf(`
	INSERT INTO %s.ftc_trade
	SET trade_no = :order_id,
		user_id = :compound_id,
		ftc_user_id = :ftc_id,
		wx_union_id = :union_id,
		trade_price = :price,
		trade_amount = :amount,
		tier_to_buy = :tier,
		billing_cycle = :cycle,
		cycle_count = :cycle_count,
		extra_days = :extra_days,
		category = :usage_type,
		payment_method = :payment_method,
		wx_app_id = wx_app_id,
		created_utc = UTC_TIMESTAMP(),
		upgrade_id = :upgrade_id,
		member_snapshot_id = :member_snapshot_id`, b.MemberDB())
}

// SelectSubsLock select an order upon webhook received
// notification.
func (b Builder) SelectSubsLock() string {
	return fmt.Sprintf(`
	SELECT trade_no AS order_id,
		user_id AS compound_id,
		ftc_user_id AS ftc_id,
		wx_union_id AS union_id,
		trade_price AS price,
		trade_amount AS amount,
		tier_to_buy AS tier,
		billing_cycle AS cycle,
		cycle_count AS cycle_count,
		extra_days AS extra_days,
		category AS usage_type,
		upgrade_id AS upgrade_id,
		payment_method AS payment_method,
		created_utc AS created_at,
		confirmed_utc AS confirmed_at
	FROM %s.ftc_trade
	WHERE trade_no = ?
	LIMIT 1
	FOR UPDATE`, b.MemberDB())
}

// Statement to update a subscription order after received notification from payment provider.
func (b Builder) ConfirmOrder() string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_trade
	SET confirmed_utc = :confirmed_at,
		start_date = :start_date,
		end_date = :end_date
	WHERE trade_no = :order_id
	LIMIT 1`, b.MemberDB())
}

func (b Builder) ConfirmationResult() string {
	return fmt.Sprintf(`
	INSERT INTO %s.confirmation_result
	SET order_id = ?,
		succeeded = ?,
		failed = ?,
		created_utc = UTC_TIMESTAMP()`, b.MemberDB())
}
