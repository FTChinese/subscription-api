package ftcpay

// StmtCreateOrder a new order. The newly created order does not have
// ConfirmedAt, StartDate and EndDate set.
const StmtCreateOrder = `
INSERT INTO premium.ftc_trade
SET trade_no = :order_id,
	user_id = :compound_id,
	ftc_user_id = :ftc_id,
	wx_union_id = :union_id,
	tier_to_buy = :tier,
	billing_cycle = :cycle,
	category = :kind,
	trade_price = :original_price,
	trade_amount = :payable_amount,
	payment_method = :payment_method,
	years_count = :years_count,
	months_count = :months_count,
	days_count = :days_count,
	wx_app_id = :wx_app_id,
	created_utc = :created_utc
`

// Split row retrieval into multiple parts to circumvent
// problematic server setting of not being able to
// retrieve too much data onece.
const colsOrderHeader = `
o.trade_no AS order_id,
o.user_id AS compound_id,
o.ftc_user_id AS ftc_id,
o.wx_union_id AS union_id,
o.tier_to_buy AS tier,
o.billing_cycle AS cycle,
o.category AS kind,
o.trade_price AS original_price,
o.trade_amount AS payable_amount
`

const colsOrderTail = `
o.payment_method,
o.years_count,
o.months_count,
o.days_count,
o.wx_app_id,
o.created_utc,
o.confirmed_utc,
o.start_date,
o.end_date
`

const StmtOrderCols = `
SELECT 
` + colsOrderHeader + `,
` + colsOrderTail

// StmtOrderHeader retrieves the first half of an order
// to circumvent the problem of oversize packet transfer between
// db and server.
const StmtOrderHeader = `
SELECT ` + colsOrderHeader + `
FROM premium.ftc_trade AS o
WHERE trade_no = ?
LIMIT 1
`

const StmtOrderTail = `
SELECT ` + colsOrderTail + `
FROM premium.ftc_trade AS o
WHERE trade_no = ?
LIMIT 1
`

const StmtSelectOrder = StmtOrderCols + `
FROM premium.ftc_trade AS o
WHERE trade_no = ?
LIMIT 1
`

// StmtLockOrder locks a row to retrieve onl the minimal fields
// used to determine whether the order is confirmed.
// This is a hack when we cannot load all fields in one shot
// caused by a mysterious MySQL error.
const StmtLockOrder = `
SELECT trade_no AS order_id,
	confirmed_utc,
	start_date,
	end_date
FROM premium.ftc_trade
WHERE trade_no = ?
LIMIT 1
FOR UPDATE`

// StmtConfirmOrder build SQL to set an order confirmed.
const StmtConfirmOrder = `
UPDATE premium.ftc_trade
SET confirmed_utc = :confirmed_utc,
	start_date = :start_date,
	end_date = :end_date
WHERE trade_no = :order_id
LIMIT 1`

const StmtListOrders = StmtOrderCols + `
FROM premium.ftc_trade AS o
WHERE FIND_IN_SET(o.user_id, ?) > 0
ORDER BY o.created_utc DESC
LIMIT ? OFFSET ?`

const StmtCountOrders = `
SELECT COUNT(*) AS row_count
FROM premium.ftc_trade
WHERE FIND_IN_SET(user_id, ?) > 0`
