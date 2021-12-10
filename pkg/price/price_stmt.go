package price

const StmtCreatePrice = `
INSERT INTO subs_product.price
SET id = :price_id,
	tier = :tier,
	cycle = :cycle,
	is_active = :is_active,
	archived = :archived,
	currency = :currency,
	kind = :kind,
	live_mode = :live_mode,
	nickname = :nickname,
	period_count = :period_count,
	product_id = :product_id,
	stripe_price_id = :stripe_price_id,
	title = :title,
	unit_amount = :unit_amount,
	start_utc = :start_utc,
	end_utc = :end_utc,
	created_utc = :created_utc
`

const StmtUpdatePrice = `
UPDATE subs_product.price
SET nickname = :nickname,
	stripe_price_id = :stripe_price_id,
	title = :title
WHERE id =:price_id
LIMIT 1
`

// StmtActivatePrice turns a price to activated state.
// Used together with the above statement to ensure that
// uniqueness of edition under a product.
const StmtActivatePrice = `
UPDATE subs_product.price
SET is_active = :is_active
WHERE id =:price_id
LIMIT 1
`

// StmtDeactivateSiblingPrices flags all sibling prices of
// specified price_id to inactive so that we could ensure a
// product won't have duplicate active price of the same edition.
// We call prices are siblings when then meet the following requirements simultaneously:
// * belonging to the same product;
// * being the same edition;
// * fall into the same kind;
// * in the same mode.
// Archived prices should be ignored.
const StmtDeactivateSiblingPrices = `
UPDATE subs_product.price
SET is_active = FALSE
WHERE product_id = :product_id
	AND tier = :tier
	AND cycle = :cycle
	AND kind = :kind
	AND live_mode = :live_mode
	AND id != :price_id
	AND archived = FALSE
`

const StmtArchivePrice = `
UPDATE subs_product.price
SET archived = :archived,
	is_active = :is_active,
	discount_list = NULL
WHERE id = :price_id
LIMIT 1
`

const ColPaywallPrice = `
SELECT p.id AS price_id,
	p.tier,
	p.cycle,
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
	p.start_utc,
	p.end_utc,
	p.created_utc
FROM subs_product.price AS p
`

const StmtListPricesOfProduct = ColPaywallPrice + `
WHERE p.product_id = ?
	AND p.live_mode = ?
	AND p.archived = FALSE
ORDER BY p.is_active DESC, p.cycle DESC, p.created_utc DESC
`
