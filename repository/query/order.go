package query

import "fmt"

// Statement to insert a subscription order.
func (b Builder) InsertSubs() string {
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
		client_type = :client_type,
		client_version = :client_version,
		user_ip_bin = INET6_ATON(:user_ip),
		user_agent = :user_agent`, b.MemberDB())
}

// SelectSubsPrice retrieves an order's price when payment
// provider send confirmation notice.
func (b Builder) SelectSubsPrice() string {
	return fmt.Sprintf(`
	SELECT trade_price AS :price,
		trade_amount AS :amount,
		confirmed_utc IS NOT NULL AS :is_confirmed
	FROM %s.ftc_trade
	WHERE trade_no = ?
	LIMIT 1`, b.MemberDB())
}

func (b Builder) FtcUserOrderBilling() string {
	return fmt.Sprintf(`
	SELECT trade_price AS listPrice,
		trade_amount AS netPrice,
		confirmed_utc IS NOT NULL AS isConfirmed
	FROM %s.ftc_trade
	WHERE trade_no = ?
		AND ftc_user_id = ?
	LIMIT 1`, b.MemberDB())
}

// SelectSubsLock select an order upon webhook received
// notification.
func (b Builder) SelectSubsLock() string {
	return fmt.Sprintf(`
	SELECT trade_no AS :order_id,
		user_id AS :compound_id,
		ftc_user_id AS :ftc_id,
		wx_union_id AS :union_id,
		trade_price AS :price,
		trade_amount AS :amount,
		tier_to_buy AS :tier,
		billing_cycle AS :cycle,
		cycle_count AS :cycle_count,
		extra_days AS :extra_days,
		category AS :usage_type,
		payment_method AS :payment_method,
		created_utc AS :created_at,
		confirmed_utc AS :confirmed_at,
		confirmed_utc IS NOT NULL AS :is_confirmed
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
	WHERE trade_no = :id
	LIMIT 1`, b.MemberDB())
}

func (b Builder) BalanceSource() string {
	return fmt.Sprintf(`
	SELECT s.trade_no AS orderId,
		s.trade_amount AS amount,
		CASE s.trade_subs
			WHEN 0 THEN s.start_date
			WHEN 10 THEN IF(
				s.billing_cycle = 'month', 
				DATE(DATE_ADD(FROM_UNIXTIME(s.trade_end), INTERVAL -1 MONTH)), 
				DATE(DATE_ADD(FROM_UNIXTIME(s.trade_end), INTERVAL -1 YEAR))
			)
			WHEN 100 THEN DATE(DATE_ADD(FROM_UNIXTIME(s.trade_end), INTERVAL -1 YEAR))
		 END AS startDate,
		IF(s.end_date IS NOT NULL, s.end_date, DATE(FROM_UNIXTIME(s.trade_end))) AS endDate
	FROM %s.ftc_trade AS s
		LEFT JOIN %s.upgrade AS u
		ON s.last_upgrade_id = u.id
	WHERE s.user_id IN (?, ?)
		AND (s.tier_to_buy = 'standard' OR s.trade_subs = 10)
		AND (
			s.end_date > UTC_DATE() OR
			s.trade_end > UNIX_TIMESTAMP()
		)
		AND (s.confirmed_utc IS NOT NULL OR s.trade_end != 0)
		AND u.confirmed_utc IS NULL
 	ORDER BY start_date ASC
	FOR UPDATE`, b.MemberDB(), b.MemberDB())
}

func (b Builder) SetLastUpgradeID() string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_trade
	SET last_upgrade_id = ?
	WHERE FIND_IN_SET(trade_no, ?) > 0`, b.MemberDB())
}

func (b Builder) InsertUpgrade() string {
	return fmt.Sprintf(`
	INSERT INTO %s.upgrade
	SET id = :up_id,
		order_id = :order_id,
		balance = :balance,
		source_id = :source_id,
		created_utc = UTC_TIMESTAMP(),
		member_id = :member_id,
		ftc_id = :ftc_id,
		wx_union_id = :union_id,
		product_tier = :tier
		cycle = :cycle,
		expire_date = :expire_date`,
		b.MemberDB(),
	)
}

func (b Builder) SelectUpgrade() string {
	return fmt.Sprintf(`
	SELECT id,
		balance,
		source_id AS sourceId,
		order_id AS orderId,
		created_utc AS createdUtc,
		confirmed_utc AS confirmedUtc
	FROM %s.upgrade
	WHERE order_id = ?
	LIMIT 1`, b.MemberDB())
}

func (b Builder) ConfirmUpgrade() string {
	return fmt.Sprintf(`
	UPDATE %s.upgrade
	SET confirmed_utc = UTC_TIMESTAMP()
	WHERE id = ?`, b.MemberDB())
}

func (b Builder) ConfirmationResult() string {
	return fmt.Sprintf(`
	INSERT INTO %s.confirmation_result
	SET order_id = ?,
		succeeded = ?,
		failed = ?,
		created_utc = UTC_TIMESTAMP()`, b.MemberDB())
}
