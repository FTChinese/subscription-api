package query

import "fmt"

type MemberCol string

const (
	MemberColCompoundID MemberCol = "vip_id"
	MemberColUnionID              = "vip_id_alias"
)

// SelectMemberLock builds statement to select a membership
// in a transaction. The where clause might be `vip_id`
// or `vip_id_alias` depending on user's current login method
// and account linking status.
// If user logged-in with ftc-only account, retrieve membership
// by `vip_id` column;
// If user logged-in with wechat-only account, retrieve it
// by `vip_id_alias` column, which should be wechat's
// union id;
// If user account is linked, use `vip_id` as ftc id should
// be enough.
func (b Builder) SelectMemberLock(col MemberCol) string {
	// 10 -> standard
	// 100 -> premium
	return fmt.Sprintf(`
	SELECT id AS member_id, 
		vip_id AS compound_id,
		NULLIF(vip_id, vip_id_alias) AS ftc_id,
		vip_id_alias AS union_id,
		vip_type,
		expire_time,
		member_tier AS tier,
		billing_cycle AS cycle,
		expire_date,
		payment_method,
		stripe_subscription_id AS stripe_sub_id,
		auto_renewal,
		sub_status
	FROM %s.ftc_vip
	WHERE %s = ?
	LIMIT 1
	FOR UPDATE`, b.MemberDB(), string(col))
}

func (b Builder) InsertMember() string {
	return fmt.Sprintf(`
	INSERT INTO %s.ftc_vip
	SET id = :member_id,
		vip_id = :compound_id,
		vip_id_alias = :union_id,
		vip_type = :vip_type,
		expire_time = :expire_time,
		ftc_user_id = :ftc_id,
		wx_union_id = :union_id,
		member_tier = :tier,
		billing_cycle = :cycle,
		expire_date = :expire_date,
		payment_method = :payment_method,
		stripe_subscription_id = :stripe_sub_id,
		stripe_plan_id = :stripe_plan_id,
		auto_renewal = :auto_renewal,
		sub_status = :sub_status`, b.MemberDB())
}

// UpdateMember update an existing member for stripe pay.
// The only works fot FTC users. Wechat user are not allowed
// to use stripe since those users are mostly located in China.
func (b Builder) UpdateMember(whereCol MemberCol) string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_vip
	SET id = :member_id,
		vip_type = :vip_type,
		expire_time = :expire_time,
		member_tier = :tier,
		billing_cycle = :cycle,
		expire_date = :expire_date,
		payment_method = :payment_method,
		stripe_subscription_id = :stripe_sub_id,
		stripe_plan_id = :stripe_plan_id,
		auto_renewal = :auto_renewal,
		sub_status = :sub_status
	WHERE %s = :coumpound_id
	LIMIT 1`, b.MemberDB(), string(whereCol))
}

func (b Builder) AddMemberID(whereCol MemberCol) string {
	return fmt.Sprintf(`
	UPDATE %s.ftc_vip
	SET id = IF(id IS NULL, :id, id)
	WHERE %s = :compound_id
	LIMIT 1`, b.MemberDB(), string(whereCol))
}
