package price

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
	"strings"
)

type FtcPriceUpdateParams struct {
	Description   null.String `json:"description"`
	Nickname      null.String `json:"nickname"`
	StripePriceID string      `json:"stripePriceId"`
}

func (p FtcPriceUpdateParams) Validate() *render.ValidationError {
	return validator.New("stripePriceId").Required().Validate(p.StripePriceID)
}

// FtcPriceParams is the form data submitted to create a price.
// A new plan is always created under a certain product.
// Therefore, the input data does not have tier field.
type FtcPriceParams struct {
	CreatedBy string `json:"createdBy"`
	Edition
	FtcPriceUpdateParams
	ProductID  string  `json:"productId"`
	UnitAmount float64 `json:"unitAmount"`
}

// Validate checks whether the input data to create a new plan is valid.
// `productTier` is used to specify for which edition of product this plan is created.
// Premium product is not allowed to have a monthly pricing plan.
func (p *FtcPriceParams) Validate() *render.ValidationError {

	p.Description.String = strings.TrimSpace(p.Description.String)

	if p.Cycle == enum.CycleNull {
		return &render.ValidationError{
			Message: "Invalid cycle",
			Field:   "cycle",
			Code:    render.CodeInvalid,
		}
	}

	if p.UnitAmount <= 0 {
		return &render.ValidationError{
			Message: "Price could not be below 0",
			Field:   "price",
			Code:    render.CodeInvalid,
		}
	}

	ve := validator.New("productId").Required().Validate(p.ProductID)
	if ve != nil {
		return ve
	}

	return p.FtcPriceUpdateParams.Validate()
}

// FtcPrice contains a price's original price and promotion.
// The actual price user paid should be the original price minus
// promotion offer if promotion period is valid.
type FtcPrice struct {
	Price
	StripePriceID string           `json:"stripePriceId" db:"stripe_price_id"`
	Offers        DiscountListJSON `json:"offers" db:"discount_list"`
}

// NewFtcPrice creates a price for ftc product
func NewFtcPrice(p FtcPriceParams, live bool) FtcPrice {

	return FtcPrice{
		Price: Price{
			ID:          ids.PriceID(),
			Edition:     p.Edition,
			Active:      false,
			Archived:    false,
			Currency:    "cny",
			Description: p.Description,
			LiveMode:    live,
			Nickname:    p.Nickname,
			ProductID:   p.ProductID,
			UnitAmount:  p.UnitAmount,
			CreatedUTC:  chrono.TimeNow(),
			CreatedBy:   p.CreatedBy,
		},
		StripePriceID: p.StripePriceID,
		Offers:        make([]Discount, 0),
	}
}

func (f FtcPrice) Update(p FtcPriceUpdateParams) FtcPrice {
	f.Description = p.Description
	f.Nickname = p.Nickname
	f.StripePriceID = p.StripePriceID

	return f
}

func (f FtcPrice) Activate() FtcPrice {
	f.Active = true
	return f
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

// Archive put a price into archive and no longer usable.
// No idea why I created this.
func (f FtcPrice) Archive() FtcPrice {
	f.Archived = true
	return f
}

// CheckoutItem contains a price user is trying to purchase and optional discount of this price
// applicable to this user.
type CheckoutItem struct {
	Price Price    `json:"price"`
	Offer Discount `json:"offer"` // Optional
}

// Verify checks if the price and offer match after retrieved from
// db separately.
func (i CheckoutItem) Verify(live bool) error {
	// If the discount does not
	if i.Price.ID != i.Offer.PriceID {
		return errors.New("the price and discount does not match")
	}

	if i.Price.LiveMode != live {
		return fmt.Errorf("price defined in %s environment cannot be used in %s environment", ids.GetBoolKey(i.Price.LiveMode), ids.GetBoolKey(live))
	}

	return nil
}
