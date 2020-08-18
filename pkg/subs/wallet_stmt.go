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
WHERE o.user_id IN (?, ?)
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

const insertProration = `
INSERT INTO %s.proration
SET order_id = :order_id,
	balance = :balance,
	created_utc = :created_at,
	consumed_utc = :consumed_at,
	upgrade_id = :upgrade_id`

func StmtSaveProration(db config.SubsDB) string {
	return fmt.Sprintf(insertProration)
}

const prorationUsed = `
UPDATE %s.proration
SET consumed_utc = UTC_TIMESTAMP()
WHERE upgrade_id = ?`

func StmtProrationUsed(db config.SubsDB) string {
	return fmt.Sprintf(prorationUsed, db)
}

const selectProration = `
SELECT order_id,
	balance,
	created_utc AS created_at,
	consumed_utc AS consumed_at,
	upgrade_id
FROM %s.proration
WHERE upgrade_id = ?`

func StmtListProration(db config.SubsDB) string {
	return fmt.Sprintf(selectProration, db)
}
