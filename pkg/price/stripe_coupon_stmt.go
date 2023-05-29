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
SELECT c.id,
	c.amount_off,
	c.created,
	c.currency,
	c.duration,
	c.end_utc,
	c.live_mode,
	c.display_name AS name,
	c.price_id,
	c.start_utc,
	c.current_status AS status,
	c.updated_utc
FROM subs_product.stripe_coupon AS c
`

const StmtRetrieveCoupon = colSelectCoupon + `
WHERE c.id = ?
	AND c.live_mode = ?
LIMIT 1
`

// StmtPriceAllCoupons retrieves all coupons of a price, regardless of
// whether they are active or not.
// Used by CMS.
const StmtPriceAllCoupons = colSelectCoupon + `
WHERE c.price_id = ?
	AND c.live_mode = ?
ORDER BY c.updated_utc DESC
`

// StmtPriceActiveCoupons retrieve all active coupons
// of a price, regardless of whether the price is
// active on paywall or not.
// Used by user-facing apps.
const StmtPriceActiveCoupons = colSelectCoupon + `
WHERE c.price_id = ?
	AND c.live_mode = ?
	AND c.current_status = 'active'
ORDER BY c.amount_off DESC
`

// StmtPaywallStripeCoupons retrieves all coupons
// of prices that are currently present on paywall.
// It is achieved by LEFT JOIN product_active_price,
// and ensure that rows from product_active_price
// is not null and source is tripe.
const StmtPaywallStripeCoupons = colSelectCoupon + `
LEFT JOIN subs_product.product_active_price AS a
	on c.price_id = a.price_id
WHERE c.live_mode = ?
	AND a.source = 'stripe'
	AND c.current_status = 'active'
	AND (c.end_utc IS NULL OR c.end_utc >= UTC_TIMESTAMP())
ORDER BY c.amount_off DESC
`

const StmtChangeCouponStatus = `
UPDATE subs_product.stripe_coupon
SET current_status = :status
WHERE id = :id
LIMIT 1
`
