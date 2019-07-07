package query

const (
	SelectFtcUser = `
	SELECT user_id AS userId,
		wx_union_id AS unionId,
		stripe_customer_id AS stripeId,
		user_name AS userName,
		email
	FROM cmstmp01.userinfo
	WHERE user_id = ?
	LIMIT 1`

	LockFtcUser = SelectFtcUser + `
	FOR UPDATE`

	SaveStripeID = `
	UPDATE cmstmp01.userinfo
	SET stripe_customer_id = ?
	WHERE user_id = ?
	LIMIT 1`
)
