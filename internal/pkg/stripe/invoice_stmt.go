package stripe

const colInsertInvoice = `
account_country = :account_country,
account_name = :account_name,
amount_due = :amount_due,
amount_paid = :amount_paid,
amount_remaining = :amount_remaining,
attempt_count = :attempt_count,
attempted = :attempted,
auto_advance = :auto_advance,
billing_reason = :billing_reason,
charge_id = :charge_id,
collection_method = :collection_method,
currency = :currency,
customer_id = :customer_id,
default_payment_method = :default_payment_method,
discount_ids = :discount_ids,
hosted_invoice_url = :hosted_invoice_url,
invoice_pdf = :invoice_pdf,
live_mode = :live_mode,
next_payment_attempt = :next_payment_attempt,
identity_number = :identity_number,
paid = :paid,
payment_intent_id = :payment_intent_id,
period_end_utc = :period_end_utc,
period_start_utc = :period_start_utc,
receipt_number = :receipt_number,
invoice_status = :invoice_status,
subscription_id = :subscription_id,
total = :total,
created = :created
`

const StmtUpsertInvoice = `
INSERT INTO premium.stripe_invoice
SET id = :id,
` + colInsertInvoice + `
ON DUPLICATE KEY UPDATE
` + colInsertInvoice + `,
updated_utc = UTC_TIMESTAMP()`

const StmtRetrieveInvoice = `
SELECT id,
	account_country,
	account_name,
	amount_due,
	amount_paid,
	amount_remaining,
	attempt_count,
	attempted,
	auto_advance,
	billing_reason,
	charge_id,
	collection_method,
	currency,
	customer_id,
	default_payment_method,
	discount_ids,
	hosted_invoice_url,
	invoice_pdf,
	live_mode,
	next_payment_attempt,
	identity_number,
	paid,
	payment_intent_id,
	period_end_utc,
	period_start_utc,
	receipt_number,
	invoice_status,
	subscription_id,
	total,
	created
FROM premium.stripe_invoice
WHERE id = ?
LIMIT 1
`
