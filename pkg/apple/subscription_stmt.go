package apple

const colsSubsBase = `
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
` + colsSubsBase + `,
	created_utc = UTC_TIMESTAMP()
ON DUPLICATE KEY UPDATE
` + colsSubsBase

const StmtLoadSubs = `
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
WHERE original_transaction_id = ?
LIMIT 1`
