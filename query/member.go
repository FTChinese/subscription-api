package query

import "fmt"

func (b Builder) SelectMemberLock() string {
	return fmt.Sprintf(`
	SELECT id, 
		vip_id AS userId,
		vip_id_alias AS unionId,
		IF(member_tier, member_tier, CASE vip_type
			WHEN 10 THEN 'standard'
			WHEN 100 THEN 'premium'
			ELSE member_tier
		END) AS tier,
		billing_cycle AS cycle,
		CASE
			WHEN expire_date IS NOT NULL THEN expire_date
			WHEN expire_time > 0 THEN DATE(FROM_UNIXTIME(expire_time))
			ELSE NULL
		END AS expireDate,
		payment_method AS paymentMethod,
		stripe_subscription_id AS stripeSubId,
		auto_renewal AS autoRenewal
	FROM %s.ftc_vip
	WHERE vip_id = ? 
		OR vip_id_alias = ?
	LIMIT 1
	FOR UPDATE`, b.MemberDB())
}

func (b Builder) InsertMember() string {
	return fmt.Sprintf(`
	INSERT INTO %s.ftc_vip
	SET id = ?,
		vip_id = ?,
		vip_id_alias = ?,
		vip_type = ?,
		expire_time = ?,
		ftc_user_id = ?,
		wx_union_id = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?,
		payment_method = ?,
		stripe_subscription_id = ?,
		stripe_plan_id = ?,
		auto_renewal = ?`, b.MemberDB())
}

// UpdateMember update an existing member for stripe pay.
// The only works fot FTC users. Wechat user are not allowed
// to use stripe since those users are mostly located in China.
func (b Builder) UpdateMember() string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_vip
	SET id = IFNULL(id, ?),
		vip_type = ?,
		expire_time = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?,
		payment_method = ?,
		stripe_subscription_id = ?,
		stripe_plan_id = ?,
		auto_renewal = ?
	WHERE vip_id IN (?, ?)
	LIMIT 1`, b.MemberDB())
}

func (b Builder) LinkFtcMember() string {
	return fmt.Sprintf(`
	UPDATE %s.userinfo
	SET member_id = ?
	WHERE user_id = ?
	LIMIT 1`, b.CmsTmp())
}

func (b Builder) LinkWxMember() string {
	return fmt.Sprintf(`
	UPDATE %s.wechat_userinfo
	SET member_id = ?
	WHERE union_id = ?
	LIMIT 1`, b.UserDB())
}
