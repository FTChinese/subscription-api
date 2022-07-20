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
current_status = :status
`

const colSelectCoupon = `
id,
amount_off,
created,
currency,
duration,
end_utc,
live_mode,
display_name AS name,
price_id,
start_utc,
current_status AS status
`

const StmtUpsertCoupon = `
INSERT INTO subs_product.stripe_coupon
SET id = :id,
` + colInsertCoupon + `
ON DUPLICATE KEY UPDATE
	updated_utc = UTC_TIMESTAMP(),
` + colInsertCoupon

const StmtRetrieveCoupon = `
SELECT ` + colSelectCoupon + `
FROM subs_product.stripe_coupon
WHERE id = ?
LIMIT 1
`

// StmtPriceCoupons retrieves all coupons of a price, regardless of
// whether they are active or not.
// When used in CMS, we need to ignore the status.
const StmtPriceCoupons = `
SELECT ` + colSelectCoupon + `
FROM subs_product.stripe_coupon
WHERE price_id = ?
ORDER BY amount_off DESC
`

// StmtActiveCouponsOfPrices retrieves coupons of the specified prices.
// Used to build paywall since we want to expose only the active ones.
const StmtActiveCouponsOfPrices = `
SELECT ` + colSelectCoupon + `
FROM subs_product.stripe_coupon
WHERE FIND_IN_SET(price_id, ?) > 0
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
