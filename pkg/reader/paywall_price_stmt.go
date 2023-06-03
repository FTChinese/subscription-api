package reader

// StmtSetPriceOffers updates price's discount list column.
const StmtSetPriceOffers = `
UPDATE subs_product.price
SET discount_list = :discount_list
WHERE id = :price_id
LIMIT 1
`

const colPriceWithOffers = `
SELECT p.id AS price_id,
	p.tier AS tier,
	p.cycle AS cycle,
	p.is_active AS is_active,
	p.archived AS archived,
	p.currency AS currency,
	p.kind AS kind,
	p.live_mode AS live_mode,
	p.nickname AS nickname,
	p.period_count AS period_count,
	p.product_id AS product_id,
	IFNULL(stripe_price_id, '') AS stripe_price_id,
	p.title AS title,
	p.unit_amount AS unit_amount,
	p.discount_list AS discount_list,
	p.start_utc AS start_utc,
	p.end_utc AS end_utc,
	p.created_utc AS created_utc
FROM subs_product.price AS p
`

// StmtSelectPaywallPrice retrieves a row from plan table, regardless of archived or not.
const StmtSelectPaywallPrice = colPriceWithOffers + `
WHERE p.id = ?
	AND p.live_mode = ?
LIMIT 1
`

// StmtListProductPrices retrieves all prices under a product,
// whether they are active or not, as long as not archived.
const StmtListProductPrices = colPriceWithOffers + `
WHERE p.product_id = ?
	AND p.live_mode = ?
	AND p.archived = FALSE
ORDER BY p.is_active DESC, p.cycle DESC, p.created_utc DESC
`

// StmtListPaywallPrice retrieves all active recurring prices
// under products, which are put on paywall.
// You get all such prices by left joining the paywall_product
// table and filter out those nullable product id.
// NOTE we don't need an extra table to record which prices
// are put on paywall.
// Deprecated. Use StmtListActivePrices.
const StmtListPaywallPrice = colPriceWithOffers + `
LEFT JOIN subs_product.paywall_product_v4 AS active_prod
	ON p.product_id = active_prod.product_id
WHERE active_prod.product_id IS NOT NULL
	AND p.is_active = TRUE
	AND p.archived = FALSE
	AND p.kind = 'recurring'
	AND p.live_mode = ?
ORDER BY p.cycle DESC
`

// StmtListActivePrices finds all prices set to active
// on paywall.
const StmtListActivePrices = colPriceWithOffers + `
RIGHT JOIN subs_product.product_active_price AS a
	ON p.id = a.price_id
WHERE p.live_mode = ?
	AND p.archived = FALSE
	AND a.source = 'ftc'
ORDER BY p.cycle DESC
`
