package pw

const StmtCreateProduct = `
INSERT INTO subs_product.product
SET id = :product_id,
	description = :description,
	heading = :heading,
	is_active = :is_active,
	live_mode = :live_mode,
    small_print = :small_print,
	tier = :tier,
    created_by = :created_by,
    created_utc = :created_utc,
    updated_utc = :updated_utc
`

const colProduct = `
SELECT id AS product_id,
	description,
	heading,
	is_active,
	live_mode,
	small_print,
	tier,
	created_by,
	created_utc,
	updated_utc
FROM subs_product.product
`

// StmtRetrieveProduct refer to paywall_product table.
const StmtRetrieveProduct = colProduct + `
WHERE id = ?
	AND live_mode = ?
LIMIT 1
`

const StmtListProducts = colProduct + `
WHERE live_mode = ?
ORDER BY is_active DESC, tier ASC, created_utc DESC
`

const StmtUpdateProduct = `
UPDATE subs_product.product
SET heading = :heading,
    description = :description,
    small_print = :small_print,
    updated_utc = :updated_utc
WHERE id = :product_id
LIMIT 1
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
