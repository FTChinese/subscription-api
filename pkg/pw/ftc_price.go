package pw

import "github.com/FTChinese/subscription-api/pkg/price"

// FtcPrice contains a price's original price and promotion.
// The actually price user paid should be the original price minus
// promotion offer if promotion period is valid.
type FtcPrice struct {
	Original       price.Price    `json:"original"`
	PromotionOffer price.Discount `json:"promotionOffer"`
}
