package stripe

import "github.com/FTChinese/subscription-api/pkg/price"

const StmtUpsertPrice = `
INSERT INTO subs_product.stripe_price
SET id = :id,
	price_data = :price_data,
	created_utc = UTC_TIMESTAMP()
ON DUPLICATE KEY UPDATE
	price_data = :price_data,
	updated_utc = UTC_TIMESTAMP()
`

const StmtRetrievePrice = `
SELECT id,
	price_data
FROM subs_product.stripe_price
WHERE id = ?
`

// PriceRow is used to save Stripe price in ftc's database
// so that we could avoid hitting Stripe API as much as possible.
type PriceRow struct {
	ID   string      `db:"id"`
	Data PriceColumn `db:"price_data"`
}

func NewPriceRow(p price.StripePrice) PriceRow {
	return PriceRow{
		ID: p.ID,
		Data: PriceColumn{
			StripePrice: p,
		},
	}
}
