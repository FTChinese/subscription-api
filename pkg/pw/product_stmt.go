package pw

const StmtPaywallProducts = `
SELECT prod.id AS product_id,
	prod.tier,
	prod.heading,
	prod.description,
	prod.small_print,
	prod.created_by AS created_by,
    prod.created_utc AS created_utc,
    prod.updated_utc AS updated_utc
FROM subs_product.paywall_product AS pp
	LEFT JOIN subs_product.product AS prod
	ON pp.product_id = prod.id
WHERE prod.id IS NOT NULL
ORDER BY prod.tier ASC`

const StmtCreateProduct = `
INSERT INTO subs_product.product
SET id = :product_id,
    tier = :tier,
    heading = :heading,
    description = :description,
    small_print = :small_print,
    created_by = :created_by,
    created_utc = :created_utc,
    updated_utc = :updated_utc
`
