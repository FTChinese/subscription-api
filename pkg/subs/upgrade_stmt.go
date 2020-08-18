package subs

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/config"
)

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

func StmtSaveUpgradeBalance(db config.SubsDB) string {
	return fmt.Sprintf(insertUpgradeSchema, db)
}

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

func StmtUpgradeBalance(db config.SubsDB) string {
	return fmt.Sprintf(selectUpgradeSchema, db)
}
