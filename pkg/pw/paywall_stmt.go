package pw

const StmtInsertPaywallDoc = `
INSERT INTO subs_product.paywall_doc
SET daily_banner = :daily_banner,
	promo_banner = :promo_banner,
	live_mode = :live_mode,
	created_utc = :created_utc,
	created_by = :created_by
`

const StmtRetrievePaywallDoc = `
SELECT id,
	daily_banner,
	promo_banner,
	live_mode,
	created_utc
FROM subs_product.paywall_doc
WHERE live_mode = ?
ORDER BY id
LIMIT 1`
