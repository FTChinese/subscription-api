package query

import "fmt"

// ProratedOrders finds all orders whose end date is later
// than today, and calculate how much left for each order
// up to current point.
func (b Builder) UnusedOrders() string {
	return fmt.Sprintf(`
	SELECT trade_no AS orderId,
		trade_amount AS netPrice,
		CASE trade_subs
			WHEN 0 THEN start_date
			WHEN 10 THEN IF(
				billing_cycle = 'month', 
				DATE(DATE_ADD(FROM_UNIXTIME(trade_end), INTERVAL -1 MONTH)), 
				DATE(DATE_ADD(FROM_UNIXTIME(trade_end), INTERVAL -1 YEAR))
			)
			WHEN 100 THEN DATE(DATE_ADD(FROM_UNIXTIME(trade_end), INTERVAL -1 YEAR))
		 END AS startDate,
		IF(end_date IS NOT NULL, end_date, DATE(FROM_UNIXTIME(trade_end))) AS endDate
	FROM %s.ftc_trade
	WHERE user_id IN (?, ?)
		AND (tier_to_buy = 'standard' OR trade_subs = 10)
		AND (confirmed_utc IS NOT NULL OR trade_end != 0)
		AND upgrade_target IS NULL
		AND (
			end_date > UTC_DATE() OR
			trade_end > UNIX_TIMESTAMP()
		)
 	ORDER BY start_date ASC`, b.MemberDB())
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
		upgrade_source = ?,
		upgrade_balance = ?,
		payment_method = ?,
		wx_app_id = ?,
		created_utc = UTC_TIMESTAMP(),
		client_type = ?,
		client_version = ?,
		user_ip_bin = INET6_ATON(?),
		user_agent = ?`, b.MemberDB())
}

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
		upgrade_source AS upgradeSource,
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

func (b Builder) Prorated() string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_trade
	SET upgrade_target = ?
	WHERE user_id IN (?, ?)
		AND FIND_IN_SET(trade_no, ?) > 0
		AND upgrade_target IS NULL`, b.MemberDB())
}
