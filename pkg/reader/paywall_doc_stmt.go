package reader

const StmtInsertPaywallDoc = `
INSERT INTO subs_product.paywall_doc
SET daily_banner = :daily_banner,
	promo_banner = :promo_banner,
	live_mode = :live_mode,
	created_utc = :created_utc
`

const StmtRetrievePaywallDoc = `
SELECT id,
	daily_banner,
	promo_banner,
	live_mode,
	created_utc
FROM subs_product.paywall_doc
WHERE live_mode = ?
ORDER BY id DESC
LIMIT 1`
