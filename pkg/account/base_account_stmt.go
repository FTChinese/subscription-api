package account

const StmtCreateFtc = `
INSERT INTO cmstmp01.userinfo
SET user_id = :ftc_id,
	wx_union_id = :wx_union_id,
	email = :email,
	password = MD5(:password),
	user_name = :user_name,
	created_utc = UTC_TIMESTAMP(),
	updated_utc = UTC_TIMESTAMP()`

// StmtCreateProfile is used when user account is created.
const StmtCreateProfile = `
INSERT INTO user_db.profile
SET user_id = :ftc_id`

const colsBaseAccount = `
SELECT 
	IFNULL(u.user_id, '')			AS ftc_id,
	u.stripe_customer_id 			AS stripe_id,
	IFNULL(u.email, '') 			AS email,
	p.mobile_phone					AS mobile_phone,
	u.user_name 					AS user_name,
	IFNULL(u.email_verified, FALSE) AS is_verified
`

// When retrieving ftc account only, use wx_union_id column.
const stmtBaseAccount = colsBaseAccount + `,
u.wx_union_id					AS wx_union_id
FROM cmstmp01.userinfo AS u
LEFT JOIN user_db.profile AS p
	ON u.user_id = p.user_id
`

// StmtLockBaseAccount retrieves account in a transaction.
// NOTE Mobile number is not retrieved in this statement.
const StmtLockBaseAccount = `
SELECT user_id 			AS ftc_id,
	wx_union_id,
	stripe_customer_id 	AS stripe_id,
	email,
	user_name,
	IFNULL(email_verified, FALSE) AS is_verified
FROM cmstmp01.userinfo
WHERE user_id = ?
LIMIT 1
FOR UPDATE`

const StmtBaseAccountByUUID = stmtBaseAccount + `
WHERE u.user_id = ?
LIMIT 1`

const StmtBaseAccountByMobile = stmtBaseAccount + `
WHERE p.mobile_phone = ?
LIMIT 1`

const StmtBaseAccountByWx = stmtBaseAccount + `
WHERE u.wx_union_id = ?
LIMIT 1`

const StmtBaseAccountOfStripe = stmtBaseAccount + `
WHERE u.stripe_customer_id = ?
LIMIT 1`

const StmtSetStripeID = `
UPDATE cmstmp01.userinfo
SET stripe_customer_id = :stripe_id
WHERE user_id = :ftc_id
LIMIT 1`
