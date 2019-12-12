package query

import "fmt"

const stmtSelectReader = `
SELECT user_id AS ftc_id,
	wx_union_id AS union_id,
	stripe_customer_id AS stripe_id,
	user_name AS user_name,
	email
FROM cmstmp01.userinfo
WHERE %s = ?
LIMIT 1
%s`

func BuildSelectReader(forStripe bool, lock bool) string {
	idCol := "user_id"
	if forStripe {
		idCol = "stripe_customer_id"
	}

	suffix := ""
	if lock {
		suffix = "FOR UPDATE"
	}
	return fmt.Sprintf(
		stmtSelectReader,
		idCol,
		suffix)
}

const SaveStripeID = `
	UPDATE cmstmp01.userinfo
	SET stripe_customer_id = :stripe_id
	WHERE user_id = :ftc_id
	LIMIT 1`
