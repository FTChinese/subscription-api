package query

const selectMembership = `
SELECT id AS sub_id, 
	vip_id AS sub_compound_id,
	NULLIF(vip_id, vip_id_alias) AS sub_ftc_id,
	vip_id_alias AS sub_union_id,
	vip_type,
	expire_time,
	member_tier AS plan_tier,
	billing_cycle AS plan_cycle,
	expire_date AS sub_expire_date,
	payment_method AS sub_pay_method,
	stripe_subscription_id AS stripe_sub_id,
	auto_renewal AS sub_auto_renew,
	sub_status,
	apple_subscription_id AS apple_sub_id
FROM %s.ftc_vip
WHERE %s = ?
LIMIT 1
%s`

const colsUpsertMembership = `
vip_type = :vip_type,
expire_time = :expire_time,
member_tier = :plan_tier,
billing_cycle = :plan_cycle,
expire_date = :sub_expire_date,
payment_method = :sub_pay_method,
stripe_subscription_id = :stripe_sub_id,
stripe_plan_id = :stripe_plan_id,
auto_renewal = :sub_auto_renew,
sub_status = :sub_status,
apple_subscription_id = :apple_sub_id`

const insertMembership = `
INSERT INTO %s.ftc_vip
SET id = :sub_id,
	vip_id = :sub_compound_id,
	vip_id_alias = :sub_union_id,
	ftc_user_id = :sub_ftc_id,
	wx_union_id = :sub_union_id,
` + colsUpsertMembership

const updateMembership = `
UPDATE %s.ftc_vip
SET id = IFNULL(id, :sub_id),
` + colsUpsertMembership + `
WHERE vip_id = :sub_compound_id
LIMIT 1`

const updateMembershipID = `
UPDATE %s.ftc_vip
SET id = IFNULL(id, :sub_id)
WHERE vip_id = :sub_compound_id
LIMIT 1`

// Delete old membership when linking to IAP.
const deleteFtcMembership = `
DELETE FROM %s.ftc_vip
WHERE  vip_id = :sub_compound_id
LIMIT 1`

const unlinkIAP = `
UPDATE %s.ftc_vip
SET apple_subscription_id = NULL
WHERE vip_id = :sub_compound_id
LIMIT 1`

const insertMemberSnapshot = `
INSERT INTO %s.member_snapshot
SET id = :snapshot_id,
	reason = :reason,
	created_utc = UTC_TIMESTAMP(),
	member_id = :sub_id,
	compound_id = :sub_compound_id,
	ftc_user_id = :sub_ftc_id,
	wx_union_id = :sub_union_id,
	tier = :plan_tier,
	cycle = :plan_cycle,
	expire_date = :sub_expire_date,
	payment_method = :sub_pay_method,
	stripe_subscription_id = :stripe_sub_id,
	stripe_plan_id = :stripe_plan_id,
	auto_renewal = :sub_auto_renew,
	sub_status = :sub_status`
