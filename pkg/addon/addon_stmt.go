package addon

const StmtCreateAddOn = `
INSERT INTO premium.ftc_addon
SET id = :id,
	tier = :tier,
	cycle = :cycle,
	cycle_count = :cycle_count,
	days_remained = :days_remained,
	carry_over_source = :carry_over_source,
	payment_method = :payment_method,
	user_compound_id = :compound_id,
	order_id = :order_id,
	plan_id = :plan_id,
	created_utc = UTC_TIMESTAMP()`

const stmtColAddOn = `
SELECT id,
	tier,
	cycle,
	cycle_count,
	days_remained,
	is_upgrade_carry_over,
	payment_method,
	user_compound_id AS compound_id,
	order_id,
	plan_id,
	created_utc,
	consumed_utc
FROM premium.ftc_addon
`

const StmtListAddOnLock = stmtColAddOn + `
WHERE FIND_IN_SET(user_compound_id, ?) > 0
	AND consumed_utc IS NULL
ORDER BY created_utc DESC
FOR UPDATE`

const StmtAddOnConsumed = `
UPDATE premium.ftc_addon
SET consumed_utc = UTC_TIMESTAMP()
WHERE FIND_IN_SET(id, ?) > 0`
