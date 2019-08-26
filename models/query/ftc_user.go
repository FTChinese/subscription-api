package query

const (
	SelectFtcUser = `
	SELECT user_id AS ftc_id,
		wx_union_id AS union_id,
		stripe_customer_id AS stripe_id,
		user_name AS user_name,
		email
	FROM cmstmp01.userinfo
	WHERE user_id = ?
	LIMIT 1`

	LockFtcUser = SelectFtcUser + `
	FOR UPDATE`

	SaveStripeID = `
	UPDATE cmstmp01.userinfo
	SET stripe_customer_id = :stripe_id
	WHERE user_id = :ftc_id
	LIMIT 1`

	SelectStripeCustomer = `
	SELECT user_id AS userId,
		wx_union_id AS unionId,
		stripe_customer_id AS stripeId,
		user_name AS userName,
		email
	FROM cmstmp01.userinfo
	WHERE stripe_customer_id = ?
	LIMIT 1`
)
