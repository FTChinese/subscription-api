package price

const colFtcPrice = `
SELECT p.id AS price_id,
	p.tier,
	p.cycle,
	p.is_active AS is_active,
	p.archived AS archived,
	p.currency AS currency,
	p.description AS description,
	p.live_mode AS live_mode,
	p.nickname AS nickname,
	p.product_id AS product_id,
	p.price AS unit_amount,
	IFNULL(stripe_price_id, '') AS stripe_price_id,
	p.created_utc AS created_utc,
	p.created_by AS created_by,
	p.discount_list AS discount_list
FROM subs_product.plan AS p
`

// StmtFtcPrice retrieves a row from plan table that is not archived.
const StmtFtcPrice = colFtcPrice + `
WHERE p.id = ?
	AND live_mode = ?
LIMIT 1
`

// StmtListPaywallPrice retrieves all active prices
// under products, which are put on paywall.
// You get all such prices by left joining the paywall_product
// table and filter out those nullable product id.
// NOTE we don't need an extra table to record which prices
// are put on paywall.
const StmtListPaywallPrice = colFtcPrice + `
LEFT JOIN subs_product.paywall_product AS active_prod
	ON p.product_id = active_prod.product_id
WHERE p.is_active = 1
	AND p.live_mode = ?
	AND active_prod.product_id IS NOT NULL
ORDER BY p.cycle DESC
`

const StmtListPricesOfProduct = colFtcPrice + `
WHERE p.product_id = ? AND p.archived = FALSE
ORDER BY p.is_active DESC, p.cycle DESC, p.created_utc DESC
`

// StmtSetPriceOffers updates price's discount list column.
const StmtSetPriceOffers = `
UPDATE subs_product.plan
SET discount_list = :discount_list
WHERE id = :price_id
LIMIT 1
`

const StmtCreatePrice = `
INSERT INTO subs_product.plan
SET id = :price_id,
	cycle = :cycle,
	tier = :tier,
	archived = :archived,
	is_active = :is_active,
	currency = :currency,
	description = :description,
	live_mode = :live_mode,
	nickname = :nickname,
	product_id = :product_id,
	price = :unit_amount,
	stripe_price_id = :stripe_price_id,
	created_utc = :created_utc,
	created_by = :created_by
`

const StmtUpdatePrice = `
UPDATE subs_product.plan
SET description = :description,
	nickname = :nickname,
	stripe_price_id = :stripe_price_id
WHERE id = :price_id
LIMIT 1
`

// StmtDeactivatePricesOfSameEdition flags all price of the
// specified edition under a product to inactive, except
// the one that will be turned to active.
// In this way we could ensure a product won't have duplicate
// active price of the same edition.
const StmtDeactivatePricesOfSameEdition = `
UPDATE subs_product.plan
SET is_active = FALSE
WHERE product_id = :product_id
	AND tier = :tier
	AND cycle = :cycle
	AND live_mode = :live_mode
	AND id != :price_id
`

// StmtActivatePrice turns a price to activated state.
// Used together with the above statement to ensure that
// uniqueness of edition under a product.
const StmtActivatePrice = `
UPDATE subs_product.plan
SET is_active = TRUE
WHERE id = :price_id
LIMIT 1`

// StmtActivatePriceUnderProduct put a price on a product's active list.
// This is not a proper approach.
// Adopt a new approach using price's is_active column.
// Deprecated. Kept only for backward compatible with v2.
const StmtActivatePriceUnderProduct = `
INSERT INTO subs_product.product_active_plans
SET plan_id = :price_id,
	product_id = :product_id,
	cycle = :cycle
ON DUPLICATE KEY UPDATE
	plan_id = :price_id
`

const StmtArchivePrice = `
UPDATE subs_product.plan
SET archived = TRUE,
	discount_list = NULL
WHERE id = :price_id
LIMIT 1`

const StmtArchivePriceDiscounts = `
UPDATE subs_product.discount
SET current_status = 'cancelled'
WHERE plan_id = :price_id`
