package reader

const colAccount = `
SELECT user_id AS ftc_id,
	wx_union_id AS union_id,
	stripe_customer_id AS stripe_id,
	user_name AS user_name,
	email
FROM cmstmp01.userinfo
`

const StmtAccountByFtcID = colAccount + `
WHERE user_id = ?
LIMIT 1`

const StmtAccountByStripeID = colAccount + `
WHERE stripe_customer_id = ?
LIMIT 1`

const StmtSetStripeID = `
UPDATE cmstmp01.userinfo
SET stripe_customer_id = :stripe_id
WHERE user_id = :ftc_id
LIMIT 1`

const StmtSandboxExists = `
SELECT EXISTS(
	SELECT *
	FROM user_db.sandbox_account
	WHERE ftc_id = ?
) AS sandboxFound`
