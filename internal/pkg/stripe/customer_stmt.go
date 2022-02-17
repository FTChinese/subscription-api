package stripe

const StmtSetCustomerID = `
UPDATE cmstmp01.userinfo
SET stripe_customer_id = :stripe_id
WHERE user_id = :ftc_id
LIMIT 1
`

const colCustomer = `
ftc_user_id = :ftc_user_id,
currency = :currency,
created = :created,
default_source_id = :default_source_id,
default_payment_method_id = :default_payment_method_id,
email = :email,
live_mode = :live_mode
`

const StmtUpsertCustomer = `
INSERT INTO premium.stripe_customer
SET id = :id,
` + colCustomer + `
ON DUPLICATE KEY UPDATE
` + colCustomer

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
