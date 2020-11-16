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

const StmtCreateSubs = `
INSERT INTO premium.apple_subscription
SET original_transaction_id = :original_transaction_id,
` + colsUpsertSubs + `,
ftc_user_id = :ftc_user_id
`

const colsSubs = `
SELECT a.environment,
	a.original_transaction_id,
	a.last_transaction_id,
	a.product_id,
	a.purchase_date_utc,
	a.expires_date_utc,
	a.tier,
	a.cycle,
	a.auto_renewal,
	a.created_utc,
	a.updated_utc,
	a.ftc_user_id
`

const StmtLoadSubs = colsSubs + `
FROM premium.apple_subscription AS a
WHERE original_transaction_id = ?
LIMIT 1`

const StmtLockSubs = StmtLoadSubs + `
FOR UPDATE`

const StmtListSubs = colsSubs + `,
	m.apple_subscription_id IS NOT NULL AS in_use
FROM premium.apple_subscription AS a
	LEFT JOIN premium.ftc_vip AS m
	ON a.original_transaction_id = m.apple_subscription_id
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
SET original_transaction_id = :original_transaction_id,
	ftc_user_id = :ftc_user_id,
	created_utc = UTC_TIMESTAMP()`

const StmtArchiveUnlink = `
INSERT INTO premium.apple_unlink_archive
SET original_transaction_id = :original_transaction_id,
	ftc_user_id = :ftc_user_id,
	created_utc = UTC_TIMESTAMP()`
