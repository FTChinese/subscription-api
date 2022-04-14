package price

const colSelectStripePrice = `
SELECT id,
	active,
	currency,
	is_intro,
	kind,
	live_mode,
	nickname,
	product_id,
	period_count,
	tier,
	unit_amount,
	start_utc,
	end_utc,
	created
FROM subs_product.stripe_price
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

const StmtRetrieveStripePrice = colSelectStripePrice + `
WHERE id = ?
LIMIT 1
`

const StmtListStripePrice = colSelectStripePrice + `
WHERE FIND_IN_SET(id, ?) > 0
ORDER BY created
`
