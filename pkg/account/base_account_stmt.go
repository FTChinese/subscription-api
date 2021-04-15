package account

const StmtCreateFtc = `
INSERT INTO cmstmp01.userinfo
SET user_id = :ftc_id,
	wx_union_id = :wx_union_id,
	stripe_customer_id = :stripe_id,
	email = :email,
	password = MD5(:password),
	user_name = :user_name,
	created_utc = UTC_TIMESTAMP(),
	updated_utc = UTC_TIMESTAMP()`

// StmtCreateProfile is used when user account is created.
const StmtCreateProfile = `
INSERT INTO user_db.profile
SET user_id = :ftc_id,
	mobile_phone = :mobile_phone`

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

const StmtBaseAccountByEmail = stmtBaseAccount + `
WHERE u.email = ?
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

// StmtLinkAccount set wechat union to ftc account row.
// This is used when both ftc and wechat accounts exists
// and user is trying to link them.
// If a Wechat user is trying to sign up a new FTC account,
// use the wechat signup workflow instead.
const StmtLinkAccount = `
UPDATE cmstmp01.userinfo
SET wx_union_id = :wx_union_id,
	updated_utc = UTC_TIMESTAMP()
WHERE user_id = :ftc_id
LIMIT 1`

// StmtUnlinkAccount sets wx_union_id to null.
const StmtUnlinkAccount = `
UPDATE cmstmp01.userinfo
SET wx_union_id = NULL
WHERE user_id = :ftc_id
	AND wx_union_id = :wx_union_id
LIMIT 1`
