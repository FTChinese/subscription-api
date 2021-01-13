package subs

const StmtCreateAddOn = `
INSERT INTO premium.ftc_addon
SET id = :id,
	tier = :tier,
	cycle = :cycle,
	cycle_count = :cycle_count,
	days_remained = :days_remained,
	payment_method = :payment_method,
	order_id = :order_id,
	compound_id = :compound_id,
	created_utc = UTC_TIMESTAMP()`

const stmtListAddOn = `
SELECT id,
	tier,
	cycle,
	cycle_count,
	days_remained,
	payment_method,
	order_id,
	compound_id,
	created_utc,
	consumed_utc
FROM premium.ftc_addon
WHERE FIND_IN_SET(compound_id, ?) > 0
	AND consumed_utc IS NULL
ORDER BY created_utc DESC`

const StmtListAddOn = stmtListAddOn + `
LIMIT ? OFFSET ?`

const StmtListAddOnLock = stmtListAddOn + `
FOR UPDATE`

const StmtAddOnConsumed = `
UPDATE premium.ftc_addon
SET consumed_utc = UTC_TIMESTAMP()
WHERE FIND_IN_SET(id, ?) > 0`
