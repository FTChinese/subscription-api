package query

import "fmt"

func (b Builder) InsertProration() string {
	return fmt.Sprintf(`
	INSERT INTO %s.proration
	SET order_id = :order_id,
		balance = :balance,
		created_utc = UTC_TIMESTAMP(),
		upgrade_id = :upgrade_id`, b.MemberDB())
}

func (b Builder) SelectProrationSource() string {
	return fmt.Sprintf(`
	SELECT o.trade_no AS order_id,
		o.trade_amount AS paid_amount,
		CASE o.trade_subs
			WHEN 0 THEN o.start_date
			WHEN 10 THEN IF(
				o.billing_cycle = 'month', 
				DATE(DATE_ADD(FROM_UNIXTIME(o.trade_end), INTERVAL -1 MONTH)), 
				DATE(DATE_ADD(FROM_UNIXTIME(o.trade_end), INTERVAL -1 YEAR))
			)
			WHEN 100 THEN DATE(DATE_ADD(FROM_UNIXTIME(o.trade_end), INTERVAL -1 YEAR))
		 END AS start_date,
		IF(o.end_date IS NOT NULL, o.end_date, DATE(FROM_UNIXTIME(o.trade_end))) AS end_date
	FROM %s.ftc_trade AS o
		LEFT JOIN %s.proration AS p
		ON o.trade_no = p.order_id
	WHERE o.user_id IN (?, ?)
		AND (o.tier_to_buy = 'standard' OR o.trade_subs = 10)
		AND (
			o.end_date > UTC_DATE() OR
			o.trade_end > UNIX_TIMESTAMP()
		)
		AND (o.confirmed_utc IS NOT NULL OR o.trade_end != 0)
		AND p.consumed_utc IS NULL
 	ORDER BY start_date ASC`, b.MemberDB(), b.MemberDB())
}

func (b Builder) InsertUpgradePlan() string {
	return fmt.Sprintf(`
	INSERT INTO %s.upgrade_plan
	SET id = :upgrade_id,
		balance = :balance,
		created_utc = UTC_TIMESTAMP(),
		plan_tier = :tier,
		plan_cycle = :cycle,
		plan_price = :price,
		plan_amount = :amount,
		plan_cycle_count = :cycle_count,
		plan_extra_days = :extra_days,
		plan_currency = :currency,
		plan_title = :title`,
		b.MemberDB(),
	)
}

func (b Builder) ProrationConsumed() string {
	return fmt.Sprintf(`
	UPDATE %s.proration
	SET consumed_utc = UTC_TIMESTAMP()
	WHERE upgrade_id = ?`, b.MemberDB())
}

func (b Builder) SelectUpgradePlan() string {
	return fmt.Sprintf(`
	SELECT id AS upgrade_id,
		balance,
		created_utc AS created_at,
		plan_tier AS tier,
		plan_cycle AS cycle,
		plan_price AS price,
		plan_amount AS amount,
		plan_cycle_count AS cycle_count,
		plan_extra_days AS extra_days,
		plan_currency AS currency,
		plan_title AS title
	FROM %s.upgrade_plan
	WHER id = ?
	LIMIT 1`, b.MemberDB())
}

func (b Builder) SelectUpgradeSource() string {
	return fmt.Sprintf(`
	SELECT order_id
		balance,
		created_utc AS created_at,
		consumed_utc AS consumed_at,
		upgrade_id
	FROM %s.proration
	WHERE upgrade_id = ?`, b.MemberDB())
}
