package model

import "fmt"

const (
	// Statement to select a row from ftc_trade table.
	selectSubs = `
	SELECT trade_no AS orderId,
		user_id AS userId,
		ftc_user_id AS ftcUserId,
		wx_union_id AS unionId,
		tier_to_buy AS tierToBuy,
		billing_cycle AS billingCycle,
		trade_price AS listPrice,
		trade_amount AS netPrice,
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
		user_id = ?,
		ftc_user_id = ?,
		wx_union_id = ?,
		trade_price = ?,
		trade_amount = ?,
		tier_to_buy = ?,
		billing_cycle = ?,
		payment_method = ?,
		created_utc = UTC_TIMESTAMP(),
		client_type = ?,
		client_version = ?,
		user_ip_bin = INET6_ATON(?),
		user_agent = ?`, env.vipDBName())
}

func (env Env) stmtSelectSubs() string {
	return fmt.Sprintf(selectSubs, env.vipDBName())
}

func (env Env) stmtSelectSubsLock() string {
	return fmt.Sprintf(selectSubsLock, env.vipDBName())
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
	LIMIT 1`, env.vipDBName())
}

// Statement to insert a new member or update an existing one after subscription order is confirmed.
func (env Env) stmtInsertMember() string {
	return fmt.Sprintf(`
	INSERT INTO %s.ftc_vip
	SET vip_id = ?,
		vip_id_alias = ?,
		ftc_user_id = ?,
		wx_union_id = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?`, env.vipDBName())
}

func (env Env) stmtUpdateMember() string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_vip
	SET ftc_user_id = ?,
		wx_union_id = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?
	WHERE vip_id = ?
		OR vip_id_alias = ?
	LIMIT 1`, env.vipDBName())
}

// Save as the above one, with more data retrieved.
func (env Env) stmtSelectMember() string {
	return fmt.Sprintf(`
	SELECT vip_id AS userId,
		vip_id_alias AS unionId,
		vip_type AS vipType,
		member_tier AS memberTier,
		billing_cycle AS billingCyce,
		expire_time AS expireTime,
		expire_date AS expireDate
	FROM %s.ftc_vip
	WHERE vip_id = ? 
		OR vip_id_alias = ?
	LIMIT 1`, env.vipDBName())
}

// Build statement select a row from ftc_vip by different criteria depending on whether user is logged-in with Wechat or not.
func (env Env) stmtSelectExpireDate() string {

	return fmt.Sprintf(`
	SELECT expire_time AS expireTime,
		expire_date AS expireDate
	FROM %s.ftc_vip
	WHERE vip_id = ? 
		OR vip_id_alias = ?
	LIMIT 1
	FOR UPDATE`, env.vipDBName())
}
