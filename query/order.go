package query

import "fmt"

func (b Builder) ProratedOrders() string {
	return fmt.Sprintf(`
	SELECT trade_no AS orderId,
		IF(
			start_date < UTC_DATE(), 
			DATEDIFF(end_date, UTC_DATE()) * trade_amount / DATEDIFF(end_date, start_date), 
			trade_amount
		) AS balance,
		start_date AS startDate,
		end_date AS endDate
	FROM %s.ftc_trade
	WHERE user_id IN (?, ?)
		AND (
			end_date > UTC_DATE() OR
			trade_end > UNIX_TIMESTAMP()
		)
		AND (confirmed_utc IS NOT NULL OR trade_end != 0)
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
		proration_source = ?,
		proration_amount = ?,
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
		proration_source AS prorationSource,
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
	SET upgrade_failed = ?
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
	SET prorated_to = ?
	WHERE user_id IN (?, ?)
		AND trade_no IN (?)
		AND prorated_to IS NULL`, b.MemberDB())
}
