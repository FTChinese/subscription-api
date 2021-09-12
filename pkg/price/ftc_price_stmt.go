package price

const colFtcPrice = `
SELECT p.id AS price_id,
	p.tier,
	p.cycle,
	p.currency AS currency,
	p.is_active AS is_active,
	p.description AS description,
	p.live_mode AS live_mode,
	p.nickname AS nickname,
	p.product_id AS product_id,
	p.source AS source,
	p.price AS unit_amount,
	p.created_utc AS created_utc,
	p.created_by AS created_by,
	p.discount_list AS discount_list
FROM subs_product.plan AS p
`

// StmtFtcPrice retrieves a row from plan table
const StmtFtcPrice = colFtcPrice + `
WHERE p.id = ?
LIMIT 1
`

const StmtLockFtcPrice = colFtcPrice + `
WHERE p.id = ?
LIMIT 1
FOR UPDATE`

// StmtListActivePrice retrieves all active prices
// for products shown on paywall.
const StmtListActivePrice = colFtcPrice + `
LEFT JOIN subs_product.paywall_product AS active_prod
	ON p.product_id = active_prod.product_id
WHERE p.is_active = TRUE
	AND p.live_mode = ?
	AND active_prod.product_id IS NOT NULL
ORDER BY p.cycle DESC
`

const StmtListPricesOfProduct = colFtcPrice + `
WHERE p.product_id = ?
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
	currency = :currency,
	is_active = :is_active,
	description = :description,
	live_mode = :live_mode,
	nickname = :nickname,
	product_id = :product_id,
	source = :source,
	price = :unit_amount,
	created_utc = :created_utc,
	created_by = :created_by
`

// StmtDeactivatePricesOfSameEdition flags all price of the
// specified edition under a product to inactive, except
// the one that will be turned to active.
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
