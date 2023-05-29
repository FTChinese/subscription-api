package price

const colSelectStripePrice = `
SELECT p.id AS id,
	p.active AS active,
	p.currency AS currency,
	p.is_intro AS is_intro,
	p.kind AS kind,
	p.live_mode as live_mode,
	p.nickname as nickname,
	a.price_id IS NOT NULL AS on_paywall,
	p.product_id AS product_id,
	p.period_count AS period_count,
	p.tier AS tier,
	p.unit_amount AS unit_amount,
	p.start_utc AS start_utc,
	p.end_utc AS end_utc,
	p.created AS created
FROM subs_product.stripe_price AS p
`

// StmtStripePrice finds a stripe price by id.
// By left joining the product_active_price table,
// we could know whether a price is present on paywall.
const StmtStripePrice = colSelectStripePrice + `
LEFT JOIN subs_product.product_active_price AS a
	ON p.id = a.price_id
WHERE p.id = ?
	AND p.live_mode = ?
	AND (a.source = 'stripe' OR a.source IS NULL)
LIMIT 1
`

// StmtStripePagedPrices retrieves a list of stripe prices with pagination.
// By left joining the product_active_price table,
// we know whether each price is present on paywall.
// The OR clause in WHERE is important to exclude
// ftc prices.
const StmtStripePagedPrices = colSelectStripePrice + `
LEFT JOIN subs_product.product_active_price AS a
	ON p.id = a.price_id
WHERE p.live_mode = ?
	AND (a.source = 'stripe' OR a.source IS NULL)
ORDER BY p.updated_utc DESC
LIMIT ? OFFSET ?
`

// StmtCOuntStripePrice count how many entries there are for stripe price.
const StmtCountStripePrice = `
SELECT COUNT(*)
FROM subs_product.stripe_price
WHERE live_mode = ?
`

// StmtStripeActivePrices retrieves all stripe prices
// that only present in product_active_price table.
const StmtStripeActivePrices = colSelectStripePrice + `
RIGHT JOIN subs_product.product_active_price AS a
	ON p.id = a.price_id
WHERE p.live_mode = ?
	AND a.source = 'stripe'
ORDER BY created
`

const colInsertStripePrice = `
active = :active,
currency = :currency,
is_intro = :is_intro,
kind = :kind,
live_mode = :live_mode,
nickname = :nickname,
product_id = :product_id,
period_count = :period_count,
tier = :tier,
unit_amount = :unit_amount,
start_utc = :start_utc,
end_utc = :end_utc,
created = :created
`

const StmtUpsertStripePrice = `
INSERT INTO subs_product.stripe_price
SET id = :id,
` + colInsertStripePrice + `
ON DUPLICATE KEY UPDATE
	updated_utc = UTC_TIMESTAMP(),
` + colInsertStripePrice
