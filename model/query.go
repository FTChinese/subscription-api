package model

import "fmt"

var memberTable = "premium"

var (
	insertSubs = fmt.Sprintf(`
	INSERT INTO %s.ftc_trade
	SET trade_no = ?,
		trade_price = ?,
		trade_amount = ?,
		user_id = ?,
		login_method = ?,
		tier_to_buy = ?,
		billing_cycle = ?,
		payment_method = ?,
		is_renewal = ?,
		created_utc = UTC_TIMESTAMP(),
		client_type = ?,
		client_version = ?,
		user_ip_bin = INET6_ATON(?),
		user_agent = ?`, memberTable)

	selectSubs = fmt.Sprintf(`
	SELECT user_id AS userId,
		trade_no AS orderId,
		trade_price AS price,
		trade_amount AS charged,
		login_method AS loginMethod,
		tier_to_buy AS tierToBuy,
		billing_cycle AS billingCycle,
		payment_method AS paymentMethod,
		created_utc AS createdAt,
		confirmed_utc AS confirmedAt
	FROM %s.ftc_trade
	WHERE trade_no = ?
	LIMIT 1`, memberTable)

	selectSubsLock = fmt.Sprintf(`%s
	FOR UPDATE`, selectSubs)

	updateSubs = fmt.Sprintf(`
	UPDATE %s.ftc_trade
	SET is_renewal = ?,
		confirmed_utc = ?,
		start_date = ?,
		end_date = ?
	WHERE trade_no = ?
	LIMIT 1`, memberTable)

	insertMember = fmt.Sprintf(`
	INSERT INTO %s.ftc_vip
	SET vip_id = ?,
		vip_id_alias = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?
	ON DUPLICATE KEY UPDATE
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?`, memberTable)
)
