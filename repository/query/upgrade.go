package query

const selectBalanceSource = `
SELECT o.trade_no AS order_id,
	o.trade_amount AS charged_amount,
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
ORDER BY start_date ASC`

const insertProration = `
INSERT INTO %s.proration
SET order_id = :order_id,
	balance = :balance,
	created_utc = :created_at,
	consumed_utc = :consumed_at,
	upgrade_id = :upgrade_id`

const prorationUsed = `
UPDATE %s.proration
SET consumed_utc = UTC_TIMESTAMP()
WHERE upgrade_id = ?`

const selectProration = `
SELECT order_id,
	balance,
	created_utc AS created_at,
	consumed_utc AS consumed_at,
	upgrade_id
FROM %s.proration
WHERE upgrade_id = ?`

const insertUpgradeSchema = `
INSERT INTO %s.upgrade_plan
SET id = :upgrade_id,
	balance = :balance,
	created_utc = UTC_TIMESTAMP(),
	plan_tier = :plan_tier,
	plan_cycle = :plan_cycle,
	plan_price = :plan_price,
	plan_amount = :plan_amount,
	plan_currency = :plan_currency`

const selectUpgradeSchema = `
SELECT id AS upgrade_id,
	balance,
	created_utc AS created_at,
	plan_tier AS plan_tier,
	plan_cycle AS plan_cycle,
	plan_price AS plan_price,
	plan_amount AS plan_amount,
	plan_currency AS plan_currency
FROM %s.upgrade_plan
WHERE id = ?
LIMIT 1`
