package price

// FtcPrice contains a price's original price and promotion.
// The actually price user paid should be the original price minus
// promotion offer if promotion period is valid.
type FtcPrice struct {
	Original       Price    `json:"original"`
	PromotionOffer Discount `json:"promotionOffer"`
}
