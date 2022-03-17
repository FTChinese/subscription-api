package stripe

const colUpsertPI = `
amount = :amount,
amount_received = :amount_received,
canceled_at = :canceled_at,
cancellation_reason = :cancellation_reason,
client_secret = :client_secret,
created = :created,
currency = :currency,
customer_id = :customer_id,
invoice_id = :invoice_id,
live_mode = :live_mode,
next_action = :next_action,
payment_method_id = :payment_method_id,
payment_method_types = :payment_method_types,
receipt_email = :receipt_email,
setup_future_usage = :setup_future_usage,
intent_status = :intent_status
`
const StmtUpsertPaymentIntent = `
INSERT INTO premium.stripe_payment_intent
SET id = :id,
` + colUpsertPI + `
ON DUPLICATE KEY UPDATE
` + colUpsertPI
