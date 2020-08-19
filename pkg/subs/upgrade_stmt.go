package subs

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/config"
)

// save the total balance when creating an upgrade order.
const insertUpgradeSchema = `
INSERT INTO %s.upgrade_plan
SET id = :upgrade_id,
	balance = :balance,
	created_utc = :created_utc,
	order_id = :order_id`

func StmtSaveUpgradeBalance(db config.SubsDB) string {
	return fmt.Sprintf(insertUpgradeSchema, db)
}

// Insert a ProratedOrderSchema.
// consumed_utc is required only when upgrading is free.
const insertProratedOrder = `
INSERT INTO %s.proration
SET order_id = :order_id,
	balance = :balance,
	created_utc = :created_utc,
	consumed_utc = :consumed_utc,
	upgrade_id = :upgrade_id`

func StmtSaveProratedOrder(db config.SubsDB) string {
	return fmt.Sprintf(insertProratedOrder, db)
}

// Flags all prorated orders as used for an upgrade operation.
const proratedOrdersUsed = `
UPDATE %s.proration AS p
LEFT JOIN upgrade_plan AS u
ON p.upgrade_id = u.id
SET p.consumed_utc = UTC_TIMESTAMP()
WHERE u.order_id = ?`

func StmtProratedOrdersUsed(db config.SubsDB) string {
	return fmt.Sprintf(proratedOrdersUsed, db)
}

// retrieves an upgrade's balance bu upgrade id.
const selectUpgradeBalance = `
SELECT id AS upgrade_id,
	balance,
	created_utc,
	order_id
FROM %s.upgrade_plan
WHERE order_id = ?
LIMIT 1`

func StmtUpgradeBalance(db config.SubsDB) string {
	return fmt.Sprintf(selectUpgradeBalance, db)
}

const selectProratedOrders = `
SELECT order_id,
	balance,
	created_utc,
	consumed_utc,
	upgrade_id
FROM %s.proration
WHERE upgrade_id = ?`

func StmtListProratedOrders(db config.SubsDB) string {
	return fmt.Sprintf(selectProratedOrders, db)
}
