package pw

const colUpsertProduct = `
description = :description,
heading = :heading,
introductory = :introductory,
small_print = :small_print
`

const StmtCreateProduct = `
INSERT INTO subs_product.product
SET id = :product_id,
	is_active = :is_active,
	live_mode = :live_mode,
	tier = :tier,
    created_by = :created_by,
    created_utc = :created_utc,
` + colUpsertProduct

const StmtUpdateProduct = `
UPDATE subs_product.product
SET ` + colUpsertProduct + `
WHERE id = :product_id
LIMIT 1
`

const colProduct = `
SELECT prod.id AS product_id,
	prod.description,
	prod.heading,
	prod.is_active,
	prod.live_mode,
	prod.small_print,
	prod.tier,
	prod.introductory,
	prod.created_by AS created_by,
    prod.created_utc AS created_utc,
    prod.updated_utc AS updated_utc
`

const fromTableProduct = `
FROM subs_product.product AS prod
`

// StmtRetrieveProduct refer to paywall_product table.
const StmtRetrieveProduct = colProduct + fromTableProduct + `
WHERE id = ?
	AND live_mode = ?
LIMIT 1
`

const StmtListProducts = colProduct + fromTableProduct + `
WHERE live_mode = ?
ORDER BY is_active DESC, tier ASC, created_utc DESC
`

const StmtDeactivateSiblingProducts = `
UPDATE subs_product.product
SET is_active = FALSE
WHERE tier = :tier
	AND live_mode = :live_mode
	AND id != :product_id
`

const StmtActivateProduct = `
UPDATE subs_product.product
SET is_active = TRUE
WHERE id = :product_id
LIMIT 1
`

// StmtPutProductOnPaywallLegacy could only put live mode product on paywall.
// Kept for backward compatible for API < 4
const StmtPutProductOnPaywallLegacy = `
INSERT INTO subs_product.paywall_product
SET product_id = :product_id,
	tier = :tier
ON DUPLICATE KEY UPDATE
	product_id = :product_id`

const StmtPutProductOnPaywall = `
INSERT INTO subs_product.paywall_product_v4
SET product_id = :product_id,
	tier = :tier,
	live_mode = :live_mode
ON DUPLICATE KEY UPDATE
	product_id = :product_id`
