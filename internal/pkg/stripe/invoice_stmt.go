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
payment_intent_id = :payment_intent_id,
period_end_utc = :period_end_utc,
period_start_utc = :period_start_utc,
receipt_number = :receipt_number,
invoice_status = :invoice_status,
total = :total,
updated_utc = :updated_utc
`

const StmtUpsertInvoice = `
INSERT INTO premium.stripe_invoice
SET id = :id,
` + colInsertInvoice + `,
created_utc = :created_utc
ON DUPLICATE KEY UPDATE
` + colInsertInvoice
