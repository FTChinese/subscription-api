package invoice

const StmtCreateInvoice = `
INSERT INTO premium.ftc_invoice
SET id = :id,
	tier = :tier,
	cycle = :cycle,
	cycle_count = :cycle_count,
	trial_days = :trial_days,
	order_kind = :order_kind,
	addon_days = :addon_days,
	addon_source = :addon_source,
	payment_method = :payment_method,
	user_compound_id = :compound_id,
	order_id = :order_id,
	price_id = :price_id,
	created_utc = UTC_TIMESTAMP(),
	consumed_utc = :consumed_utc,
	start_utc = :start_utc,
	end_utc = :end_utc`

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
	tier,
	cycle,
	cycle_count,
	trial_days,
	order_kind,
	addon_days,
	addon_source,
	payment_method,
	user_compound_id AS compound_id,
	order_id,
	price_id,
	created_utc,
	consumed_utc,
	start_utc,
	end_utc
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
const StmtInvoiceConsumed = `
UPDATE premium.ftc_invoice
SET consumed_utc = :consumed_utc,
	start_utc = :start_utc,
	end_utc = :end_utc
WHERE id = :id`
