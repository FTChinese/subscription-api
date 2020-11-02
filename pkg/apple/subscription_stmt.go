package apple

const colsUpsertSubs = `
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
` + colsUpsertSubs + `,
	created_utc = UTC_TIMESTAMP()
ON DUPLICATE KEY  UPDATE
` + colsUpsertSubs

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
	updated_utc,
	ftc_user_id
FROM premium.apple_subscription
`

const StmtLoadSubs = colsSubs + `
WHERE original_transaction_id = ?
LIMIT 1`

const StmtLockSubs = StmtLoadSubs + `
FOR UPDATE`

const StmtListSubs = colsSubs + `
WHERE ftc_user_id = ?
ORDER BY expires_date_utc DESC
LIMIT ? OFFSET ?`

const StmtCountSubs = `
SELECT COUNT(*) AS row_count
FROM premium.apple_subscription
WHERE ftc_user_id = ?`

const StmtLinkSubs = `
UPDATE premium.apple_subscription
SET ftc_user_id = :ftc_user_id
WHERE original_transaction_id = :original_transaction_id
LIMIT 1`

const StmtUnlinkSubs = `
UPDATE premium.apple_subscription
SET ftc_user_id = NULL
WHERE original_transaction_id = :original_transaction_id
LIMIT 1`

const StmtArchiveLinkCheat = `
INSERT INTO premium.apple_cheat
SET original_transaction_id = :original_transaction_id
	ftc_user_id = :ftc_user_id,
	created_utc = UTC_TIMESTAMP()`

const StmtArchiveUnlink = `
INSERT INTO premium.apple_unlink_archive
SET original_transaction_id = :original_transaction_id
	ftc_user_id = :ftc_user_id,
	created_utc = UTC_TIMESTAMP()`
