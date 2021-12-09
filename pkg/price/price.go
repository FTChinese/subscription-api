package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
	"strings"
)

type UpdateParams struct {
	Description   null.String `json:"description"`
	Nickname      null.String `json:"nickname"`
	StripePriceID string      `json:"stripePriceId"`
}

func (p UpdateParams) Validate() *render.ValidationError {
	return validator.New("stripePriceId").Required().Validate(p.StripePriceID)
}

// CreationParams is the form data submitted to create a price.
// A new plan is always created under a certain product.
// Therefore, the input data does not have tier field.
type CreationParams struct {
	CreatedBy string `json:"createdBy"`
	Kind      Kind   `json:"kind"`
	Edition
	UpdateParams
	ProductID  string  `json:"productId"`
	UnitAmount float64 `json:"unitAmount"`
}

// Validate checks whether the input data to create a new plan is valid.
// `productTier` is used to specify for which edition of product this plan is created.
// Premium product is not allowed to have a monthly pricing plan.
func (p *CreationParams) Validate() *render.ValidationError {

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

	return p.UpdateParams.Validate()
}

// Price presents the price of a price. It unified prices coming
// from various source, e.g., FTC in-house or Stripe API.
type Price struct {
	ID string `json:"id" db:"price_id"`
	Edition
	Active        bool        `json:"active" db:"is_active"`
	Archived      bool        `json:"archived" db:"archived"`
	Currency      Currency    `json:"currency" db:"currency"`
	Description   null.String `json:"description" db:"description"`
	Kind          Kind        `json:"kind" db:"kind"`
	LiveMode      bool        `json:"liveMode" db:"live_mode"`
	Nickname      null.String `json:"nickname" db:"nickname"`
	ProductID     string      `json:"productId" db:"product_id"`
	StripePriceID string      `json:"stripePriceId" db:"stripe_price_id"`
	UnitAmount    float64     `json:"unitAmount" db:"unit_amount"`
	CreatedUTC    chrono.Time `json:"createdUtc" db:"created_utc"`
	CreatedBy     string      `json:"createdBy" db:"created_by"` // Use-facing client should ignore this field.
}

func New(p CreationParams, live bool) Price {
	return Price{
		ID:            ids.PriceID(),
		Edition:       p.Edition,
		Active:        false,
		Archived:      false,
		Currency:      "cny",
		Description:   p.Description,
		Kind:          p.Kind,
		LiveMode:      live,
		Nickname:      p.Nickname,
		ProductID:     p.ProductID,
		StripePriceID: p.StripePriceID,
		UnitAmount:    p.UnitAmount,
		CreatedUTC:    chrono.TimeNow(),
		CreatedBy:     p.CreatedBy,
	}
}

func (p Price) Update(params UpdateParams) Price {
	p.Description = params.Description
	p.Nickname = params.Nickname
	p.StripePriceID = params.StripePriceID

	return p
}

func (p Price) Activate() Price {
	p.Active = true

	return p
}

// Archive put a price into archive and no longer usable.
// No idea why I created this.
func (p Price) Archive() Price {
	p.Archived = true
	return p
}

// DailyCost calculates the daily average price depending on the cycles.
// Deprecated. Moved to client.
func (p Price) DailyCost() DailyCost {
	switch p.Cycle {
	case enum.CycleYear:
		return NewDailyCostOfYear(p.UnitAmount)

	case enum.CycleMonth:
		return NewDailyCostOfMonth(p.UnitAmount)
	}

	return DailyCost{}
}
