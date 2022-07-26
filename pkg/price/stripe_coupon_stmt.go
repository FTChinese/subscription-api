package price

const colInsertCoupon = `
amount_off = :amount_off,
created = :created,
currency = :currency,
duration = :duration,
end_utc = :end_utc,
live_mode = :live_mode,
display_name = :name,
price_id = :price_id,
redeem_by = :redeem_by,
start_utc = :start_utc,
current_status = :status,
updated_utc = :updated_utc
`

const StmtUpsertCoupon = `
INSERT INTO subs_product.stripe_coupon
SET id = :id,
` + colInsertCoupon + `
ON DUPLICATE KEY UPDATE
` + colInsertCoupon

const colSelectCoupon = `
SELECT id,
	amount_off,
	created,
	currency,
	duration,
	end_utc,
	live_mode,
	display_name AS name,
	price_id,
	start_utc,
	current_status AS status,
	updated_utc
FROM subs_product.stripe_coupon
`

const StmtRetrieveCoupon = colSelectCoupon + `
WHERE id = ?
	AND live_mode = ?
LIMIT 1
`

// StmtPriceAllCoupons retrieves all coupons of a price, regardless of
// whether they are active or not.
// Used by CMS.
const StmtPriceAllCoupons = colSelectCoupon + `
WHERE price_id = ?
	AND live_mode = ?
ORDER BY updated_utc DESC
`

// StmtPriceActiveCoupons retrieve all active coupons of a price.
// Used by user-facing apps.
const StmtPriceActiveCoupons = colSelectCoupon + `
WHERE price_id = ?
	AND live_mode = ?
	AND current_status = 'active'
ORDER BY amount_off DESC
`

// StmtActiveCouponsOfPrices retrieves coupons of the specified prices.
// Used to build paywall since we want to expose only the active ones.
const StmtActiveCouponsOfPrices = colSelectCoupon + `
WHERE FIND_IN_SET(price_id, ?) > 0
	AND live_mode = ?
	AND current_status = 'active'
	AND (end_utc IS NULL OR end_utc >= UTC_TIMESTAMP())
ORDER BY amount_off DESC
`

const StmtChangeCouponStatus = `
UPDATE subs_product.stripe_coupon
SET current_status = :status
WHERE id = :id
LIMIT 1
`
