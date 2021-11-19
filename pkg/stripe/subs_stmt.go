package stripe

const colUpsertSubs = `
tier = :tier,
cycle = :cycle,
cancel_at_utc = :cancel_at_utc,
cancel_at_period_end = :cancel_at_period_end,
canceled_utc = :canceled_utc,
current_period_end = :current_period_end,
current_period_start = :current_period_start,
customer_id = :customer_id,
default_payment_method = :default_payment_method,
price_id = :price_id,
price = :price,
latest_invoice_id = :latest_invoice_id,
live_mode = :live_mode,
start_date_utc = :start_date_utc,
ended_utc = :ended_utc,
updated_utc = :updated_utc,
subs_item_id = :subs_item_id,
sub_status = :sub_status
`

const StmtInsertSubs = `
INSERT INTO premium.stripe_subscription
SET id = :id,
` + colUpsertSubs + `,
created_utc = :created_utc,
ftc_user_id = :ftc_user_id
ON DUPLICATE KEY UPDATE
` + colUpsertSubs

const StmtRetrieveSubs = `
SELECT id,
	tier,
	cycle,
	cancel_at_utc,
	cancel_at_period_end,
	current_period_end,
	current_period_start,
	customer_id,
	default_payment_method,
	ended_utc,
	ftc_user_id,
	latest_invoice_id,
	live_mode,
	price,
	price_id,
	start_date_utc,
	subs_item_id,
	sub_status,
	created_utc,
	updated_utc
FROM premium.stripe_subscription
WHERE id = ?
LIMIT 1`

const StmtSubsExists = `
SELECT EXISTS(
	SELECT *
	FROM premium.stripe_subscription
	WHERE ID = ?
) AS already_exists`
