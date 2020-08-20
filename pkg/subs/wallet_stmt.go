package subs

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/config"
)

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
WHERE FIND_IN_SET(o.user_id, ?) > 0
	AND (o.tier_to_buy = 'standard' OR o.trade_subs = 10)
	AND (
		o.end_date > UTC_DATE() OR
		o.trade_end > UNIX_TIMESTAMP()
	)
	AND (o.confirmed_utc IS NOT NULL OR o.trade_end != 0)
	AND p.consumed_utc IS NULL
ORDER BY start_date ASC`

func StmtBalanceSource(db config.SubsDB) string {
	return fmt.Sprintf(selectBalanceSource, db, db)
}

// Insert a ProratedOrderSchema.
// consumed_utc is required only when upgrading is free.
const insertProratedOrder = `
INSERT INTO %s.proration
SET order_id = :order_id,
	balance = :balance,
	created_utc = :created_utc,
	consumed_utc = :consumed_utc,
	upgrade_order_id = :upgrade_order_id`

// StmtSaveProratedOrder saves an order's balance.
func StmtSaveProratedOrder(db config.SubsDB) string {
	return fmt.Sprintf(insertProratedOrder, db)
}

// Flags all prorated orders as used for an upgrade order.
const proratedOrdersUsed = `
UPDATE %s.proration
SET consumed_utc = UTC_TIMESTAMP()
WHERE upgrade_order_id = ?`

func StmtProratedOrdersUsed(db config.SubsDB) string {
	return fmt.Sprintf(proratedOrdersUsed, db)
}

const selectProratedOrders = `
SELECT order_id,
	balance,
	created_utc,
	consumed_utc,
	upgrade_order_id
FROM %s.proration
WHERE upgrade_order_id = ?`

// StmtListProratedOrders retrieves all prorated orders for an
// upgrade order.
func StmtListProratedOrders(db config.SubsDB) string {
	return fmt.Sprintf(selectProratedOrders, db)
}
