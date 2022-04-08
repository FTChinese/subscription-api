package stripe

import "github.com/FTChinese/subscription-api/pkg/price"

const colUpsertPrice = `
id = :id
price_data = :price_data
`
const StmtUpsertPrice = `
INSERT INTO subs_product.stripe_price
SET ` + colUpsertPrice + `
ON DUPLICATE KEY UPDATE
` + colUpsertPrice

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
