package price

const StmtCreatePlan = `
INSERT INTO subs_product.plan
SET id = :id,
	tier = :tier,
	cycle = :cycle,
	is_active = :active,
	archived = :archived,
	currency = :currency,
	description = :description,
	live_mode = :live_mode,
	nickname = :nickname,
	product_id = :product_id,
	stripe_price_id = :stripe_price_id,
	price = :unit_amount,
	created_by = "",
	created_utc = :created_utc
`
