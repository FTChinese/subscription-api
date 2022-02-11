package stripe

const colUpsertPaymentMethod = `
customer_id = :customer_id,
kind = :kind,
card_details = :card_details,
created = :created,
live_mode = :live_mode
`

const StmtInsertPaymentMethod = `
INSERT INTO premium.stripe_payment_method
SET id = :id,
` + colUpsertPaymentMethod + `
ON DUPLICATE KEY UPDATE
` + colUpsertPaymentMethod

const colPaymentMethod = `
SELECT id,
	customer_id,
	kind,
	card_details,
	created,
	live_mode
FROM premium.stripe_payment_method
`

const StmtRetrievePaymentMethod = colPaymentMethod + `
WHERE id = ?
LIMIT 1
`
