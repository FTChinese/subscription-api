package stripe

const colInsertInvoice = `
auto_advance = :auto_advance,
charge_id = :charge_id,
collection_method = :collection_method,
currency = :currency,
customer_id = :customer_id,
default_payment_method = :default_payment_method,
hosted_invoice_url = :hosted_invoice_url,
paid = :paid,
live_mode = :live_mode,
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
	auto_advance,
	charge_id,
	collection_method,
	currency,
	customer_id,
	default_payment_method,
	hosted_invoice_url,
	paid,
	live_mode,
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
