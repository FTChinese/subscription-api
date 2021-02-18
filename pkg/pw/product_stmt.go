package pw

const StmtPaywallProducts = `
SELECT prod.id AS product_id,
	prod.tier,
	prod.heading,
	prod.description,
	prod.small_print
FROM subs_product.paywall_product AS pp
	LEFT JOIN subs_product.price AS prod
	ON pp.product_id = prod.id
WHERE prod.id IS NOT NULL
ORDER BY prod.tier ASC`
