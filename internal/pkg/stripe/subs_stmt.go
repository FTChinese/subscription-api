package stripe

// The payment intent id only exists when request comes from client.
// Wehbook does not expand the latest_invoice.payment_intent field.
const colUpsertSubs = `
tier = :tier,
cycle = :cycle,
cancel_at_utc = :cancel_at_utc,
cancel_at_period_end = :cancel_at_period_end,
canceled_utc = :canceled_utc,
current_period_end = :current_period_end,
current_period_start = :current_period_start,
customer_id = :customer_id,
default_payment_method = :default_payment_method_id,
ended_utc = :ended_utc,
ftc_user_id = :ftc_user_id,
items = :items,
latest_invoice_id = :latest_invoice_id,
live_mode = :live_mode,
start_date_utc = :start_date_utc,
sub_status = :sub_status,
created = :created
`

const colClientUpsertSubs = colUpsertSubs + `,
payment_intent_id = :payment_intent_id
`

const StmtUpsertSubsExpanded = `
INSERT INTO premium.stripe_subscription
SET id = :id,
` + colClientUpsertSubs + `
ON DUPLICATE KEY UPDATE
` + colClientUpsertSubs + `,
updated_utc = UTC_TIMESTAMP()
`

// StmtUpsertSubsNotExpanded handles insert/update subscription without expansion,
// which does not have payment intent id field
// and should not overwrite it.
const StmtUpsertSubsNotExpanded = `
INSERT INTO premium.stripe_subscription
SET id = :id,
` + colUpsertSubs + `
ON DUPLICATE KEY UPDATE
` + colUpsertSubs + `,
updated_utc = UTC_TIMESTAMP()
`

const StmtRetrieveSubs = `
SELECT id,
	tier,
	cycle,
	cancel_at_utc,
	cancel_at_period_end,
	current_period_end,
	current_period_start,
	customer_id,
	default_payment_method AS default_payment_method_id,
	ended_utc,
	ftc_user_id,
	items,
	latest_invoice_id,
	live_mode,
	payment_intent_id,
	start_date_utc,
	sub_status,
	created
FROM premium.stripe_subscription
WHERE id = ?
LIMIT 1`
