package pw

const colUpsertProduct = `
description = :description,
heading = :heading,
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

const colSelectProduct = `
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
const StmtRetrieveProduct = colSelectProduct + fromTableProduct + `
WHERE id = ?
	AND live_mode = ?
LIMIT 1
`

const StmtListProducts = colSelectProduct + fromTableProduct + `
WHERE live_mode = ?
ORDER BY is_active DESC, tier ASC, created_utc DESC
`

// StmtSetProductIntro sets the introductory column of product.
// This happens when:
// * Activating a one_time price;
// * Updating a one_time price which was set on a product.
const StmtSetProductIntro = `
UPDATE subs_products.product
SET introductory = :introductory
WHERE id = :product_id,
	updated_utc = :updated_utc
LIMIT 1
`

// StmtDeactivateSiblingProducts deactivate sibling products of
// the specified product.
// By sibling, we mean:
// * Product of same tier;
// * in same mode
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
