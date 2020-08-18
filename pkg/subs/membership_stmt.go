package subs

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/config"
)

const colMembership = `
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
	apple_subscription_id AS apple_sub_id,
	b2b_licence_id
FROM %s.ftc_vip`

// Retrieve membership by compound id extracted from request header.
// The request might provide ftc id or union id, or both,
// and we cannot be sure the current state account ids
// are consistent with the those in db.
// There are chances that the request provides union id
// while vip_id is ftc id and vip_id_alias is union id.
// (Chances of such case are rare).
// In such case we won't be able to find the membership
// simply querying the vip_id column.
const selectMembership = colMembership + `
WHERE ? IN (vip_id, vip_id_alias)
LIMIT 1
`

func StmtMember(db config.SubsDB) string {
	return fmt.Sprintf(selectMembership, db)
}

func StmtLockMembership(db config.SubsDB) string {
	return fmt.Sprintf(selectMembership, db) + "FOR UPDATE"
}

func StmtAppleMembership(db config.SubsDB) string {
	return fmt.Sprintf(colMembership, db) + `
	WHERE apple_subscription_id = ?
	FOR UPDATE`
}

const mUpsertSharedCols = `
expire_date = :expire_date,
payment_method = :pay_method,
stripe_subscription_id = :stripe_subs_id,
stripe_plan_id = :stripe_plan_id,
auto_renewal = :auto_renew,
sub_status = :sub_status,
apple_subscription_id = :apple_subs_id,
b2b_licence_id = :b2b_licence_id
`

const mUpsertCols = `
vip_type = :vip_type,
expire_time = :expire_time,
member_tier = :tier,
billing_cycle = :cycle,
` + mUpsertSharedCols

const stmtInsertMember = `
INSERT INTO %s.ftc_vip
SET vip_id = :compound_id,
	vip_id_alias = :union_id,
	ftc_user_id = :ftc_id,
	wx_union_id = :union_id,
` + mUpsertCols

func StmtCreateMember(db config.SubsDB) string {
	return fmt.Sprintf(stmtInsertMember, db)
}

const stmtUpdateMember = `
UPDATE %s.ftc_vip
SET ` + mUpsertCols + `
WHERE vip_id = :compound_id
LIMIT 1`

func StmtUpdateMember(db config.SubsDB) string {
	return fmt.Sprintf(stmtUpdateMember, db)
}

// Delete old membership when linking to IAP.
const stmtDeleteMember = `
DELETE FROM %s.ftc_vip
WHERE  vip_id = :compound_id
LIMIT 1`

func StmtDeleteMember(db config.SubsDB) string {
	return fmt.Sprintf(stmtDeleteMember, db)
}
