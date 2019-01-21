package model

import "fmt"

const (

	// Statement to select a row from ftc_trade table.
	selectSubs = `
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
	LIMIT 1`

	// Statement to select a row from ftc_trade table used in a transaction, for row-level table locking.
	selectSubsLock = selectSubs + `
	FOR UPDATE`
)

// Statement to create a new subscription order.
func (env Env) stmtInsertSubs() string {
	return fmt.Sprintf(`
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
		user_agent = ?`, env.memberTable)
}

func (env Env) stmtSelectSubs() string {
	return fmt.Sprintf(selectSubs, env.memberTable)
}

func (env Env) stmtSelectSubsLock() string {
	return fmt.Sprintf(selectSubsLock, env.memberTable)
}

// Statement to update a subscription order after received notification from payment provider.
func (env Env) stmtUpdateSubs() string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_trade
	SET is_renewal = ?,
		confirmed_utc = ?,
		start_date = ?,
		end_date = ?
	WHERE trade_no = ?
	LIMIT 1`, env.memberTable)
}

// Statement to insert a new member or update an existing one after subscription order is confirmed.
func (env Env) stmtInsertMember() string {
	return fmt.Sprintf(`
	INSERT INTO %s.ftc_vip
	SET vip_id = ?,
		vip_id_alias = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?
	ON DUPLICATE KEY UPDATE
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?`, env.memberTable)
}

// Build statement select a row from ftc_vip by different criteria depending on whether user is logged-in with Wechat or not.
func (env Env) stmtSelectExpLock(isWxLogin bool) string {
	whereCol := "vip_id"

	if isWxLogin {
		whereCol = "vip_id_alias"
	}

	return fmt.Sprintf(`
	SELECT expire_time AS expireTime,
		expire_date AS expireDate
	FROM %s.ftc_vip
	WHERE %s = ?
	LIMIT 1
	FOR UPDATE`, env.memberTable, whereCol)
}

// Save as the above one, with more data retrieved.
func (env Env) stmtSelectMember(isWxLogin bool) string {
	whereCol := "vip_id"

	if isWxLogin {
		whereCol = "vip_id_alias"
	}

	return fmt.Sprintf(`
	SELECT vip_id AS userId,
		vip_id_alias AS unionId,
		vip_type AS vipType,
		member_tier AS memberTier,
		billing_cycle AS billingCyce,
		expire_time AS expireTime,
		expire_date AS expireDate
	FROM %s.ftc_vip
	WHERE %s = ?
	LIMIT 1`, env.memberTable, whereCol)
}
