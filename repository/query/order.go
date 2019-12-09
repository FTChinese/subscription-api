package query

const insertClientApp = `
INSERT INTO %s.client
SET order_id = :order_id,
	client_type = :client_type,
	client_version = :client_version,
	user_ip = INET6_ATON(:user_ip),
	user_agent = :user_agent`

const insertOrder = `
INSERT INTO %s.ftc_trade
	SET trade_no = :order_id,
		user_id = :sub_compound_id,
		ftc_user_id = :sub_ftc_id,
		wx_union_id = :sub_union_id,
		trade_price = :price,
		trade_amount = :amount,
		tier_to_buy = :sub_tier,
		billing_cycle = :sub_cycle,
		cycle_count = :cycle_count,
		extra_days = :extra_days,
		category = :usage_type,
		payment_method = :payment_method,
		wx_app_id = wx_app_id,
		created_utc = UTC_TIMESTAMP(),
		upgrade_id = :upgrade_id,
		member_snapshot_id = :member_snapshot_id`

const selectOrder = `
SELECT trade_no AS order_id,
	user_id AS sub_compound_id,
	ftc_user_id AS sub_ftc_id,
	wx_union_id AS sub_union_id,
	trade_price AS price,
	trade_amount AS amount,
	tier_to_buy AS sub_tier,
	billing_cycle AS sub_cycle,
	cycle_count AS cycle_count,
	extra_days AS extra_days,
	category AS usage_type,
	upgrade_id AS upgrade_id,
	payment_method AS payment_method,
	created_utc AS created_at,
	confirmed_utc AS confirmed_at
FROM %s.ftc_trade
WHERE trade_no = ?
LIMIT 1
FOR UPDATE`

const updateConfirmedOrder = `
UPDATE %s.ftc_trade
SET confirmed_utc = :confirmed_at,
	start_date = :start_date,
	end_date = :end_date
WHERE trade_no = :order_id
LIMIT 1`

const InsertConfirmationResult = `
INSERT INTO premium.confirmation_result
SET order_id = :order_id,
	succeeded = :succeeded,
	failed = :failed,
	created_utc = UTC_TIMESTAMP()`
