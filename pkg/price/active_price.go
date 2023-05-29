package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/conv"
)

type PriceSource string

const (
	PriceSourceFTC    PriceSource = "ftc"
	PriceSourceStripe PriceSource = "stripe"
)

// ActivePrice keeps track of the prices currently used under a product.
// The ID is a md5 hash of a price's various properties. It won't change
// once set. We use it to ensure there's only one active price
// of the same source, tier, cycle, kind and mode
// is available under a specific product.
//
// The price's key attributes are concatenated into a string in the format:
// <source>.<tier>.<cycle>.<kind>.<mode>
// * source: ftc or stripe
// * tier: standard or premium
// * cycle: year or month
// * kind: one_time or recurring
// * mode: live or sandbox.
//
// In theory there are a total of 2^5 entries.
// However, a premium edition usually does not have
// month and one_time options.
// See implementation of FtcPrice.uniqueFeatures
// and StripePrice.uniqueFeatures
type ActivePrice struct {
	ID         conv.HexBin `db:"id"`
	Source     PriceSource `db:"source"`
	ProductID  string      `db:"product_id"`
	PriceID    string      `db:"price_id"`
	UpdatedUTC chrono.Time `db:"updated_utc"`
}

const colActivePrice = `
source = :source,
product_id = :product_id,
price_id = :price_id,
updated_utc = :updated_utc
`

// StmtUpsertActivePrice inserts an ActivePrice into
// active price table.
const StmtUpsertActivePrice = `
INSERT INTO subs_product.product_active_price
SET id = :id,
` + colActivePrice + `
ON DUPLICATE KEY UPDATE
` + colActivePrice

// StmtRemoveActivePrice deletes a record from active price table.
// NOTE you should not use the `ID` field to perform the match
// since different prices with the same key attributes
// will have the same MD5 hash in this table.
// When you are deleting an active price, you are targeting
// the price itself, not its unique features.
const StmtRemoveActivePrice = `
DELETE 
FROM subs_product.product_active_price
WHERE source = :source
	AND price_id = :price_id
LIMIT 1
`

// Check if a price is present on paywall
const StmtIsActivePrice = `
SELECT EXISTS (
	SELECT *
	FROM subs_product.product_active_price
	WHERE price_id = ?
		AND source = ?
) AS active
`
