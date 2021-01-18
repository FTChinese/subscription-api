package subs

// Saves a new order. The newly created order does not have
// ConfirmedAt, StartDate and EndDate set.
const StmtInsertOrder = `
INSERT INTO premium.ftc_trade
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
	category = :kind,
	payment_method = :payment_method,
	wx_app_id = :wx_app_id,
	created_utc = UTC_TIMESTAMP()`

const StmtOrderCols = `
SELECT o.trade_no AS order_id,
	o.trade_price AS price,
	o.trade_amount AS charged_amount,
	o.user_id AS compound_id,
	o.ftc_user_id AS ftc_id,
	o.wx_union_id AS union_id,
	IFNULL(o.plan_id, '') AS plan_id,
	o.discount_id,
	o.tier_to_buy AS tier,
	o.billing_cycle AS cycle,
	o.category AS kind,
	o.payment_method,
	o.wx_app_id,
	o.created_utc,
	o.confirmed_utc,
	o.start_date,
	o.end_date
`

const StmtSelectOrder = StmtOrderCols + `
FROM premium.ftc_trade AS o
WHERE trade_no = ?
LIMIT 1
`

const StmtLockOrder = `
SELECT trade_no AS order_id,
	confirmed_utc
FROM premium.ftc_trade
WHERE trade_no = ?
LIMIT 1
FOR UPDATE`

// StmtOrderHeader retrieves the first half of an order
// to circumvent the problem of oversize packet transfer between
// db and server.
const StmtOrderHeader = `
SELECT trade_no AS order_id,
	trade_price AS price,
	trade_amount AS charged_amount,
	user_id AS compound_id,
	ftc_user_id AS ftc_id,
	wx_union_id AS union_id,
	IFNULL(plan_id, '') AS plan_id,
	discount_id,
	tier_to_buy AS tier,
	billing_cycle AS cycle
FROM premium.ftc_trade
WHERE trade_no = ?
LIMIT 1`

const StmtOrderTail = `
SELECT category AS kind,
	payment_method,
	wx_app_id,
	created_utc,
	confirmed_utc,
	start_date,
	end_date
FROM premium.ftc_trade
WHERE trade_no = ?
LIMIT 1`

// StmtConfirmOrder build SQL to set an order confirmed.
const StmtConfirmOrder = `
UPDATE premium.ftc_trade
SET confirmed_utc = :confirmed_utc,
	start_date = :start_date,
	end_date = :end_date
WHERE trade_no = :order_id
LIMIT 1`
