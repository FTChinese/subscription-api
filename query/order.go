package query

import "fmt"

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

func (b Builder) SetUpgradeIDOnSource() string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_trade
	SET last_upgrade_id = ?
	WHERE FIND_IN_SET(trade_no, ?) > 0`, b.MemberDB())
}

func (b Builder) InsertUpgrade() string {
	return fmt.Sprintf(`
	INSERT INTO %s.upgrade
	SET id = ?,
		order_id = ?,
		balance = ?,
		source_id = ?,
		created_utc = UTC_TIMESTAMP(),
		member_id = ?,
		cycle = ?,
		expire_date = ?,
		ftc_id = ?,
		wx_union_id = ?,
		product_tier = ?`,
		b.MemberDB(),
	)
}

func (b Builder) ConfirmUpgrade() string {
	return fmt.Sprintf(`
	UPDATE %s.upgrade
	SET confirmed_utc = UTC_TIMESTAMP()
	WHERE id = ?`, b.MemberDB())
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
	SELECT trade_no AS orderId,
		user_id AS userId,
		ftc_user_id AS ftcUserId,
		wx_union_id AS unionId,
		trade_price AS listPrice,
		trade_amount AS amount,
		tier_to_buy AS tier,
		billing_cycle AS Cycle,
		cycle_count AS cycleCount,
		extra_days AS extraDays,
		category AS usageType,
		payment_method AS paymentMethod,
		created_utc AS createdAt,
		confirmed_utc AS confirmedAt,
		confirmed_utc IS NOT NULL AS confirmed
	FROM %s.ftc_trade
	WHERE trade_no = ?
	LIMIT 1
	FOR UPDATE`, b.MemberDB())
}

func (b Builder) UpgradeFailure() string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_trade
	SET result = ?,
		failure_reason = ?
		confirmed_utc = UTC_TIMESTAMP()
	WHERE trade_no = ?
	LIMIT 1`, b.MemberDB())
}

// Statement to update a subscription order after received notification from payment provider.
func (b Builder) ConfirmSubs() string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_trade
	SET confirmed_utc = ?,
		result = 'success',
		start_date = ?,
		end_date = ?
	WHERE trade_no = ?
	LIMIT 1`, b.MemberDB())
}
