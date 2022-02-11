package stripe

const StmtSetCustomerID = `
UPDATE cmstmp01.userinfo
SET stripe_customer_id = :stripe_id
WHERE user_id = :ftc_id
LIMIT 1
`

const colCusBase = `
currency = :currency,
created = :created,
default_source_id = :default_source_id,
default_payment_method_id = :default_payment_method_id,
email = :email,
live_mode = :live_mode
`

const StmtInsertCustomer = `
INSERT INTO premium.stripe_customer
SET id = :id,
	ftc_user_id = :ftc_user_id,
` + colCusBase

// StmtUpdateCustomer updates an existing stripe customer.
// Avoid writing to ftc_user_id column.
const StmtUpdateCustomer = `
UPDATE premium.stripe_customer
SET ` + colCusBase + `
WHERE id = :id
LIMIT 1
`

const StmtRetrieveCustomer = `
SELECT id,
	IFNULL(ftc_user_id, '') AS ftc_user_id,
	currency,
	created,
	default_source_id,
	default_payment_method_id,
	email,
	live_mode
FROM premium.stripe_customer
WHERE id = ?
LIMIT 1`
