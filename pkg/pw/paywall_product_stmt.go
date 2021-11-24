package pw

// StmtPaywallProducts retrieves products put on paywall.
// Current active products under paywall is specified by
// paywall_product table. By joining it with product table
// you will retrieve only those put on paywall.
const StmtPaywallProducts = `
SELECT prod.id AS product_id,
	prod.tier,
	prod.heading,
	prod.description,
	prod.small_print,
	prod.created_by AS created_by,
    prod.created_utc AS created_utc,
    prod.updated_utc AS updated_utc
FROM subs_product.paywall_product_v4 AS pp
	LEFT JOIN subs_product.product AS prod
		ON pp.product_id = prod.id
WHERE prod.id IS NOT NULL
	AND pp.live_mode = ?
ORDER BY prod.tier ASC`
