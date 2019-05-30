package query

import "fmt"

// ProratedOrders finds all orders whose end date is later
// than today, and calculate how much left for each order
// up to current point.
func (b Builder) UnusedOrders() string {
	return fmt.Sprintf(`
	SELECT s.trade_no AS orderId,
		s.trade_amount AS netPrice,
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
		LEFT JOIN %s.upgrade_premium AS t
		ON s.trade_no = t.source_order_id
	WHERE s.user_id IN (?, ?)
		AND (s.tier_to_buy = 'standard' OR s.trade_subs = 10)
		AND (
			s.end_date > UTC_DATE() OR
			s.trade_end > UNIX_TIMESTAMP()
		)
		AND (s.confirmed_utc IS NOT NULL OR s.trade_end != 0)
		AND t.confirmed_utc IS NULL
 	ORDER BY start_date ASC
	FOR UPDATE`, b.MemberDB(), b.MemberDB())
}

func (b Builder) InsertUpgradeSource() string {
	return fmt.Sprintf(`
	INSERT INTO %s.upgrade_premium (
		source_order_id,
		target_order_id,
		created_utc)
	SELECT trade_no, 
		?, 
		UTC_TIMESTAMP()
	FROM %s.ftc_trade
	WHERE FIND_IN_SET(trade_no, ?) > 0`,
		b.MemberDB(),
		b.MemberDB(),
	)
}

func (b Builder) SelectUpgradeSource() string {
	return fmt.Sprintf(`
	SELECT source_order_id
	FROM %s.upgrade_premium
	WHERE target_order_id = ?
		AND confirmed_utc IS NULL
	FOR UPDATE`, b.MemberDB())
}

// Statement to insert a subscription order.
func (b Builder) InsertSubs() string {
	return fmt.Sprintf(`
	INSERT INTO %s.ftc_trade
	SET trade_no = ?,
		user_id = ?,
		ftc_user_id = ?,
		wx_union_id = ?,
		trade_price = ?,
		trade_amount = ?,
		tier_to_buy = ?,
		billing_cycle = ?,
		cycle_count = ?,
		extra_days = ?,
		category = ?,
		upgrade_balance = ?,
		payment_method = ?,
		wx_app_id = ?,
		created_utc = UTC_TIMESTAMP(),
		client_type = ?,
		client_version = ?,
		user_ip_bin = INET6_ATON(?),
		user_agent = ?`, b.MemberDB())
}

// SelectSubsPrice retrieves an order's price when payment
// provider send confirmation notice.
func (b Builder) SelectSubsPrice() string {
	return fmt.Sprintf(`
	SELECT trade_price AS listPrice,
		trade_amount AS netPrice,
		confirmed_utc IS NOT NULL AS isConfirmed
	FROM %s.ftc_trade
	WHERE trade_no = ?
	LIMIT 1`, b.MemberDB())
}

func (b Builder) SelectSubsLock() string {
	return fmt.Sprintf(`
	SELECT trade_no AS orderId,
		user_id AS userId,
		ftc_user_id AS ftcUserId,
		wx_union_id AS unionId,
		trade_price AS listPrice,
		trade_amount AS netPrice,
		tier_to_buy AS tier,
		billing_cycle AS Cycle,
		cycle_count AS cycleCount,
		extra_days AS extraDays,
		category AS cateogry,
		payment_method AS paymentMethod,
		confirmed_utc AS confirmedAt,
		confirmed_utc IS NOT NULL AS isConfirmed
	FROM %s.ftc_trade
	WHERE trade_no = ?
	LIMIT 1
	FOR UPDATE`, b.MemberDB())
}

func (b Builder) InvalidUpgrade() string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_trade
	SET upgrade_failed = ?,
		confirmed_utc = UTC_TIMESTAMP()
	WHERE trade_no = ?
	LIMIT 1`, b.MemberDB())
}

func (b Builder) ConfirmUpgradeSource() string {
	return fmt.Sprintf(`
	UPDATE %s.upgrade_premium
		SET confirmed_utc = UTC_TIMESTAMP()
	WHERE target_order_id = ?`, b.MemberDB())
}

// Statement to update a subscription order after received notification from payment provider.
func (b Builder) ConfirmSubs() string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_trade
	SET confirmed_utc = ?,
		start_date = ?,
		end_date = ?
	WHERE trade_no = ?
	LIMIT 1`, b.MemberDB())
}

// Deprecate
func (b Builder) Prorated() string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_trade
	SET upgrade_target = ?
	WHERE user_id IN (?, ?)
		AND FIND_IN_SET(trade_no, ?) > 0
		AND upgrade_target IS NULL`, b.MemberDB())
}
