package price

// StmtCreateDiscount creates a new discount for a plan.
const StmtCreateDiscount = `
INSERT INTO subs_product.discount
SET id = :discount_id,
	live_mode = :live_mode,
	current_status = :current_status,
	description = :discount_desc,
	kind = :kind,
	override_period = :override_period,
	percent = :percent,
	price_off = :price_off,
	plan_id = :price_id,
	recurring = :recurring,
    created_utc = :created_utc,
	start_utc = :start_utc,
    end_utc = :end_utc,
    created_by = :created_by
`

const StmtUpdateDiscount = `
UPDATE subs_product.discount
SET current_status = :current_status
WHERE id = :discount_id
LIMIT 1`

const colDiscount = `
SELECT id AS discount_id,
	live_mode,
	current_status,
	description AS discount_desc,
	kind,
	override_period,
	percent,
	start_utc,
	end_utc,
	price_off,
	plan_id AS price_id,
	recurring,
	created_utc,
	created_by
FROM subs_product.discount
`

const StmtSelectDiscount = colDiscount + `
WHERE id = ?
`

// StmtListPriceActiveDiscounts loads a list of active discount
// under a price whose end_utc is not past yet.
// TODO: filter expired ones.
const StmtListPriceActiveDiscounts = colDiscount + `
WHERE plan_id = ?
	AND current_status = 'active'
	AND live_mode = ?
	AND (end_utc IS NULL OR end_utc >= UTC_TIMESTAMP)
ORDER BY price_off DESC`

// StmtListDiscountsOfPrice retrieves all discount under a price.
const StmtListDiscountsOfPrice = colDiscount + `
WHERE plan_id = ?
ORDER BY created_utc DESC`
