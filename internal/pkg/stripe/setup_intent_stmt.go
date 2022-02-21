package stripe

const colUpsertSetupIntent = `
cancellation_reason = :cancellation_reason,
client_secret = :client_secret,
created = :created,
customer_id = :customer_id,
live_mode = :live_mode,
next_action = :next_action,
payment_method_id = :payment_method_id,
payment_method_types = :payment_method_types,
intent_status = :intent_status,
intent_usage = :intent_usage
`

const StmtUpsertSetupIntent = `
INSERT INTO premium.stripe_setup_intent
SET id = :id,
` + colUpsertSetupIntent + `
ON DUPLICATE KEY UPDATE
` + colUpsertSetupIntent

const RetrieveSetupIntent = `
SELECT id
	cancellation_reason,
	client_secret,
	created,
	customer_id,
	live_mode,
	next_action,
	payment_method_id,
	payment_method_types,
	intent_status,
	intent_usage
FROM premium.stripe_setup_intent
WHERE id = ?
LIMIT 1
`
