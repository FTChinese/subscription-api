package subs

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/config"
)

const insertOrder = `
INSERT INTO %s.ftc_trade
SET trade_no = :order_id,
	user_id = :compound_id,
	ftc_user_id = :ftc_id,
	wx_union_id = :union_id,
	trade_price = :price,
	trade_amount = :charged_amount,
	tier_to_buy = :tier,
	billing_cycle = :cycle,
	cycle_count = :cycle_count,
	extra_days = :extra_days,
	category = :kind,
	payment_method = :payment_method,
	wx_app_id = wx_app_id,
	created_utc = UTC_TIMESTAMP()`

func StmtCreateOrder(db config.SubsDB) string {
	return fmt.Sprintf(insertOrder, db)
}

const selectOrder = `
SELECT trade_no AS order_id,
	user_id AS compound_id,
	ftc_user_id AS ftc_id,
	wx_union_id AS union_id,
	trade_price AS price,
	trade_amount AS charged_amount,
	tier_to_buy AS tier,
	billing_cycle AS cycle,
	cycle_count AS cycle_count,
	extra_days AS extra_days,
	category AS kind,
	payment_method AS payment_method,
	created_utc AS created_at,
	confirmed_utc AS confirmed_at
FROM %s.ftc_trade
WHERE trade_no = ?
LIMIT 1
FOR UPDATE`

func StmtOrder(db config.SubsDB) string {
	return fmt.Sprintf(selectOrder, db)
}

const stmtConfirmOrder = `
UPDATE %s.ftc_trade
SET confirmed_utc = :confirmed_at,
	start_date = :start_date,
	end_date = :end_date
WHERE trade_no = :order_id
LIMIT 1`

func StmtConfirmOrder(db config.SubsDB) string {
	return fmt.Sprintf(stmtConfirmOrder, db)
}
