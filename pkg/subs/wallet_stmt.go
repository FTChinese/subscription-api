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
