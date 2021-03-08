package invoice

const StmtCreateInvoice = `
INSERT INTO premium.ftc_invoice
SET id = :id,
	user_compound_id = :compound_id,
	tier = :tier,
	cycle = :cycle,
	years = :years,
	months = :months,
	extra_days = :days,
	addon_source = :addon_source,
	apple_tx_id = :apple_tx_id,
	order_id = :order_id,
	order_kind = :order_kind,
	paid_amount = :paid_amount,
	payment_method = :payment_method,
	price_id = :price_id,
	stripe_subs_id = :stripe_subs_id,
	created_utc = :created_utc,
	consumed_utc = :consumed_utc,
	start_utc = :start_utc,
	end_utc = :end_utc,
	carried_over_utc = :carried_over_utc`

// StmtCarriedOver adds current moment to all invoices whose end time is after
// so that we know that an invoice's remaining time is carried over to a new
// add-on invoice.
// This usually happens when one-time purchase user upgraded from standard to premium,
// or switched to Stripe subscription.
const StmtCarriedOver = `
UPDATE premium.ftc_invoice
SET carried_over_utc = UTC_TIMESTAMP()
WHERE FIND_IN_SET(user_compound_id, ?) > 0
	AND end_utc > UTC_TIMESTAMP()`

const StmtAddOnExistsForOrder = `
SELECT EXISTS(
	SELECT * 
	FROM premium.ftc_invoice
	WHERE order_id = ?
)`

const stmtColInvoice = `
SELECT id,
	user_compound_id AS compound_id,
	tier,
	cycle,
	years,
	months,
	extra_days AS days,
	addon_source,
	order_id,
	order_kind,
	paid_amount,
	payment_method,
	price_id,
	created_utc,
	consumed_utc,
	start_utc,
	end_utc,
	carried_over_utc
FROM premium.ftc_invoice
`

// StmtListAddOnInvoiceLock retrieves all invoices that is not consumed yet
// and order kind is add_on.
const StmtListAddOnInvoiceLock = stmtColInvoice + `
WHERE FIND_IN_SET(user_compound_id, ?) > 0
	AND consumed_utc IS NULL
	AND order_kind IS NULL
ORDER BY created_utc ASC
FOR UPDATE`

// StmtInvoiceConsumed sets a invoice's consumption time and and start end end time this invoice granted to membership access.
const StmtAddOnInvoiceConsumed = `
UPDATE premium.ftc_invoice
SET consumed_utc = :consumed_utc,
	start_utc = :start_utc,
	end_utc = :end_utc
WHERE id = :id
LIMIT 1`
