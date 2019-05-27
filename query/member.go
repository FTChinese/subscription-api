package query

import "fmt"

func (b Builder) SelectMember() string {
	return fmt.Sprintf(selectMember, b.MemberDB())
}

func (b Builder) SelectMemberLock() string {
	return fmt.Sprintf(`
	SELECT vip_id AS userId,
		vip_id_alias AS unionId,
		CASE vip_type
			WHEN 10 THEN 'standard'
			WHEN 100 THEN 'premium'
			ELSE member_tier
		END AS tier,
		billing_cycle AS cycle,
		IF(expire_time, DATE(FROM_UNIXTIME(expire_time)), expire_date) AS expireDate
	FROM %s.ftc_vip
	WHERE vip_id = ? 
		OR vip_id_alias = ?
	LIMIT 1
	FOR UPDATE`, b.MemberDB())
}

func (b Builder) UpsertMember() string {
	return fmt.Sprintf(`
	INSERT INTO %s.ftc_vip
	SET vip_id = ?,
		vip_id_alias = ?,
		ftc_user_id = ?,
		wx_union_id = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?
	ON DUPLICATE KEY UPDATE
		vip_id = ?,
		vip_id_alias = ?,
		ftc_user_id = ?,
		wx_union_id = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?`, b.MemberDB())
}
