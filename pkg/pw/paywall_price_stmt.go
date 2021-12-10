package pw

import "github.com/FTChinese/subscription-api/pkg/price"

// StmtSetPriceOffers updates price's discount list column.
const StmtSetPriceOffers = `
UPDATE subs_product.price
SET discount_list = :discount_list
WHERE id = :price_id
LIMIT 1
`

// StmtSelectPaywallPrice retrieves a row from plan table, regardless of archived or not.
const StmtSelectPaywallPrice = price.ColPaywallPrice + `
WHERE p.id = ?
	AND live_mode = ?
LIMIT 1
`

// StmtListPaywallPrice retrieves all active recurring prices
// under products, which are put on paywall.
// You get all such prices by left joining the paywall_product
// table and filter out those nullable product id.
// NOTE we don't need an extra table to record which prices
// are put on paywall.
const StmtListPaywallPrice = price.ColPaywallPrice + `
LEFT JOIN subs_product.paywall_product_v4 AS active_prod
	ON p.product_id = active_prod.product_id
WHERE active_prod.product_id IS NOT NULL
	AND p.is_active = TRUE
	AND p.archived = FALSE
	AND p.kind = 'recurring'
	AND p.live_mode = ? 
ORDER BY p.cycle DESC
`
