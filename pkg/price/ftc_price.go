package price

import "sort"

// FtcPrice contains a price's original price and promotion.
// The actually price user paid should be the original price minus
// promotion offer if promotion period is valid.
type FtcPrice struct {
	Price
	PromotionOffer Discount   `json:"promotionOffer"` // Deprecated
	Offers         []Discount `json:"offers"`
}

func (f FtcPrice) ApplicableOffer(filters []OfferKind) Discount {
	// Filter all valid discount offers.

	var filtered = make([]Discount, 0)
	for _, v := range f.Offers {
		if v.IsValid() && v.Kind.ContainedBy(filters) {
			filtered = append(filtered, v)
		}
	}

	switch len(filtered) {
	case 0:
		return Discount{}
	case 1:
		return filtered[0]
	default:
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].PriceOff.Float64 > filtered[j].PriceOff.Float64
		})

		return filtered[0]
	}
}
