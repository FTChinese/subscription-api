package apple

const colsInsertSubs = `
environment = :environment,
last_transaction_id = :last_transaction_id,
product_id = :product_id,
purchase_date_utc = :purchase_date_utc,
expires_date_utc = :expires_date_utc,
tier = :tier,
cycle = :cycle,
auto_renewal = :auto_renewal,
updated_utc = UTC_TIMESTAMP()
`

const StmtUpsertSubs = `
INSERT INTO premium.apple_subscription
SET original_transaction_id = :original_transaction_id,
` + colsInsertSubs + `,
	created_utc = UTC_TIMESTAMP()
ON DUPLICATE KEY UPDATE
` + colsInsertSubs

const colsSubs = `
SELECT environment,
	original_transaction_id,
	last_transaction_id,
	product_id,
	purchase_date_utc,
	expires_date_utc,
	tier,
	cycle,
	auto_renewal,
	created_utc,
	updated_utc
FROM premium.apple_subscription
`

const StmtLoadSubs = colsSubs + `
WHERE original_transaction_id = ?
LIMIT 1`

const StmtListSubs = colsSubs + `
ORDER BY updated_utc DESC
LIMIT ? OFFSET ?`

const StmtCountSubs = `
SELECT COUNT(*) AS row_count
FROM premium.apple_subscription`
