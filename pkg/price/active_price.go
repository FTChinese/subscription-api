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
// of the same source, tier, cycle, kind and mode.
type ActivePrice struct {
	ID         conv.HexBin `db:"id"`
	Source     PriceSource `db:"source"`
	ProductID  string      `db:"product_id"`
	PriceID    string      `db:"price_id"`
	UpdatedUTC chrono.Time `db:"updated_utc"`
}

func NewActivePriceFtc(p FtcPrice) ActivePrice {
	return ActivePrice{
		ID:         p.ActiveID().ToHexBin(),
		Source:     PriceSourceFTC,
		ProductID:  p.ProductID,
		PriceID:    p.ID,
		UpdatedUTC: chrono.TimeUTCNow(),
	}
}

const colActivePrice = `
source = :source,
product_id = :product_id,
price_id = :price_id,
updated_utc = :updated_utc
`

const StmtUpsertActivePrice = `
INSERT INTO subs_product.product_active_price
SET id = :id,
` + colActivePrice + `
ON DUPLICATE KEY UPDATE
` + colActivePrice

const StmtRemoveActivePrice = `
DELETE subs_product.product_active_price
WHERE id = ?
LIMIT 1
`
