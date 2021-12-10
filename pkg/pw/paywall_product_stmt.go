package pw

// StmtPaywallProducts retrieves products put on paywall.
// Current active products under paywall is specified by
// paywall_product table. By joining it with product table
// you will retrieve only those put on paywall.
const StmtPaywallProducts = colSelectProduct + `
FROM subs_product.paywall_product_v4 AS pp
	LEFT JOIN subs_product.product AS prod
		ON pp.product_id = prod.id
WHERE prod.id IS NOT NULL
	AND pp.live_mode = ?
ORDER BY prod.tier ASC
`
