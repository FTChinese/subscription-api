package price

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
	"strings"
)

// FtcPriceParams is the form data submitted to create a price.
// A new plan is always created under a certain product.
// Therefore, the input data does not have tier field.
type FtcPriceParams struct {
	CreatedBy string `json:"createdBy"`
	Edition
	Description null.String `json:"description"`
	LiveMode    bool        `json:"liveMode"`
	Nickname    null.String `json:"nickname"`
	Price       float64     `json:"price"`
	ProductID   string      `json:"productId"`
}

// Validate checks whether the input data to create a new plan is valid.
// `productTier` is used to specify for which edition of product this plan is created.
// Premium product is not allowed to have a monthly pricing plan.
func (p *FtcPriceParams) Validate() *render.ValidationError {

	p.Description.String = strings.TrimSpace(p.Description.String)

	ve := validator.New("productId").Required().Validate(p.ProductID)
	if ve != nil {
		return ve
	}

	if p.Price <= 0 {
		return &render.ValidationError{
			Message: "Price could not be below 0",
			Field:   "price",
			Code:    render.CodeInvalid,
		}
	}

	if p.Cycle == enum.CycleNull {
		return &render.ValidationError{
			Message: "Invalid cycle",
			Field:   "cycle",
			Code:    render.CodeInvalid,
		}
	}

	return nil
}

// NewFtcPrice creates a price for ftc product
func NewFtcPrice(p FtcPriceParams) FtcPrice {
	return FtcPrice{
		Price: Price{
			ID:          ids.PriceID(),
			Edition:     p.Edition,
			Active:      false,
			Currency:    "cny",
			Description: p.Description,
			LiveMode:    true,
			Nickname:    p.Nickname,
			ProductID:   p.ProductID,
			Source:      SourceFTC,
			UnitAmount:  p.Price,
			CreatedUTC:  chrono.TimeNow(),
			CreatedBy:   p.CreatedBy,
		},
		Offers: nil,
	}
}

// FtcPrice contains a price's original price and promotion.
// The actual price user paid should be the original price minus
// promotion offer if promotion period is valid.
type FtcPrice struct {
	Price
	PromotionOffer Discount         `json:"promotionOffer"` // Deprecated
	Offers         DiscountListJSON `json:"offers" db:"discount_list"`
}

func (f FtcPrice) SetOffers(o []Discount) FtcPrice {
	f.Offers = o
	return f
}

func (f FtcPrice) VerifyOffer(o Discount) error {
	for _, v := range f.Offers {
		if v.ID == o.ID {
			return nil
		}
	}

	return errors.New("the requested offer is not found")
}

type CheckoutItem struct {
	Price Price    `json:"price"`
	Offer Discount `json:"offer"` // Optional
}
