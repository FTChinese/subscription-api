package price

// StmtCreatePlan is used sync price table with legacy plan
// table. Do not remove this unless API v5 is removed.
const StmtCreatePlan = `
INSERT INTO subs_product.plan
SET id = :id,
	tier = :tier,
	cycle = :cycle,
	is_active = :is_active,
	archived = :archived,
	currency = :currency,
	live_mode = :live_mode,
	nickname = :nickname,
	product_id = :product_id,
	stripe_price_id = :stripe_price_id,
	description = :title,
	price = :unit_amount,
	created_by = "",
	created_utc = :created_utc
`
