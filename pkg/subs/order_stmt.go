package subs

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/config"
)

// Saves a new order. The newly created order does not have
// ConfirmedAt, StartDate and EndDate set.
const insertOrder = `
INSERT INTO %s.ftc_trade
SET trade_no = :order_id,
	trade_price = :price,
	trade_amount = :charged_amount,
	user_id = :compound_id,
	ftc_user_id = :ftc_id,
	wx_union_id = :union_id,
	plan_id = :plan_id,
	discount_id = :discount_id,
	tier_to_buy = :tier,
	billing_cycle = :cycle,
	cycle_count = :cycle_count,
	extra_days = :extra_days,
	category = :kind,
	payment_method = :payment_method,
	total_balance = :total_balance,
	wx_app_id = :wx_app_id,
	created_utc = UTC_TIMESTAMP()`

func StmtCreateOrder(db config.SubsDB) string {
	return fmt.Sprintf(insertOrder, db)
}

// Retrieves an order. This is mostly used upon confirmation.
const selectOrder = `
SELECT trade_no AS order_id,
	trade_price AS price,
	trade_amount AS charged_amount,
	user_id AS compound_id,
	ftc_user_id AS ftc_id,
	wx_union_id AS union_id,
	plan_id,
	discount_id,
	tier_to_buy AS tier,
	billing_cycle AS cycle,
	cycle_count AS cycle_count,
	extra_days AS extra_days,
	category AS kind,
	payment_method,
	total_balance,
	wx_app_id,
	created_utc,
	confirmed_utc,
	start_date,
	end_date
FROM %s.ftc_trade
WHERE trade_no = ?
LIMIT 1
FOR UPDATE`

// StmtOrder build the SQL to retrieve an order.
func StmtOrder(db config.SubsDB) string {
	return fmt.Sprintf(selectOrder, db)
}

const stmtConfirmOrder = `
UPDATE %s.ftc_trade
SET confirmed_utc = :confirmed_utc,
	start_date = :start_date,
	end_date = :end_date
WHERE trade_no = :order_id
LIMIT 1`

// StmtConfirmOrder build SQL to set an order confirmed.
func StmtConfirmOrder(db config.SubsDB) string {
	return fmt.Sprintf(stmtConfirmOrder, db)
}
