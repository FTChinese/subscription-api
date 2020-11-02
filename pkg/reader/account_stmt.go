package reader

const colAccount = `
SELECT 
	user_id 				AS ftc_id,
	wx_union_id 			AS union_id,
	stripe_customer_id 		AS stripe_id,
	email					AS email,
	user_name 				AS user_name,
	IFNULL(is_vip, FALSE) AS is_vip
FROM cmstmp01.userinfo
`

const StmtAccountByFtcID = colAccount + `
WHERE user_id = ?
LIMIT 1`

const StmtAccountByWx = `
SELECT
	IFNULL(u.user_id, '')	AS ftc_id,
	w.union_id				AS union_id,
	u.stripe_customer_id	AS stripe_id,
	IFNULL(u.email, '')		AS email,
	u.user_name				AS user_name,
	IFNULL(u.is_vip, FALSE) AS is_vip
FROM user_db.wechat_userinfo AS w
	LEFT JOIN cmstmp01.userinfo AS u
	ON w.union_id = u.wx_union_id
WHERE w.union_id = ?
LIMIT 1`

const StmtAccountByStripeID = colAccount + `
WHERE stripe_customer_id = ?
LIMIT 1`

const StmtSetStripeID = `
UPDATE cmstmp01.userinfo
SET stripe_customer_id = :stripe_id
WHERE user_id = :ftc_id
LIMIT 1`
