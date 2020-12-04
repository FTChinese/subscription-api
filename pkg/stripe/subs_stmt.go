package stripe

const colUpsertSubs = `
cancel_at_utc = :cancel_at_utc,
cancel_at_period_end = :cancel_at_period_end,
current_period_end = :current_period_end,
current_period_start = :current_period_start,
customer_id = :customer_id,
default_payment_method = :default_payment_method,
latest_invoice_id = :latest_invoice_id,
live_mode = :live_mode,
start_date_utc = :start_date_utc,
end_date_utc = :end_date_utc,
created_utc = :created_utc,
updated_utc = :updated_utc,
sub_status = :status
`

const StmtInsertSubs = `
INSERT INTO premium.stripe_subscription
SET id = :id,
` + colUpsertSubs + `,
ftc_user_id = :ftc_user_id
ON DUPLICATE KEY UPDATE
` + colUpsertSubs

const StmtRetrieveSubs = `
SELECT id,
	cancel_at_utc,
	cancel_at_period_end,
	current_period_end,
	current_period_start,
	customer_id,
	default_payment_method,
	latest_invoice_id,
	live_mode,
	start_date_utc,
	end_date_utc,
	created_utc,
	updated_utc,
	subs_status AS status,
	ftc_user_id
FROM premium.stripe_subscription
WHERE id = ?
LIMIT 1`
