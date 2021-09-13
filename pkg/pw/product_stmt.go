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

const StmtUpdateProduct = `
UPDATE subs_product.product
SET heading = :heading,
    description = :description,
    small_print = :small_print,
    updated_utc = :updated_utc
WHERE id = :product_id
LIMIT 1`

const StmtActivateProduct = `
INSERT INTO subs_product.paywall_product
SET product_id = :product_id,
	tier = :tier
ON DUPLICATE KEY UPDATE
	product_id = :product_id`

const StmtHasActivePlan = `
SELECT EXISTS (
	SELECT *
	FROM subs_product.product_active_plans
	WHERE product_id = ?
) AS has_plan`
