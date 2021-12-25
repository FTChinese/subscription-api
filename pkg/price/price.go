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
	Title         null.String        `json:"title"`
	Nickname      null.String        `json:"nickname"`
	PeriodCount   ColumnYearMonthDay `json:"periodCount"`
	StripePriceID string             `json:"stripePriceId"`
}

func (p UpdateParams) Validate() *render.ValidationError {
	return validator.New("stripePriceId").Required().Validate(p.StripePriceID)
}

// CreationParams is the form data submitted to create a price.
// A new plan is always created under a certain product.
// Therefore, the input data does not have tier field.
type CreationParams struct {
	Edition
	Kind Kind `json:"kind"`
	UpdateParams
	ProductID  string      `json:"productId"`
	StartUTC   chrono.Time `json:"startUtc"`
	EndUTC     chrono.Time `json:"endUtc"`
	UnitAmount float64     `json:"unitAmount"`
}

// Validate checks whether the input data to create a new plan is valid.
// `productTier` is used to specify for which edition of product this plan is created.
// Premium product is not allowed to have a monthly pricing plan.
func (p *CreationParams) Validate() *render.ValidationError {

	if p.UnitAmount <= 0 {
		return &render.ValidationError{
			Message: "Price could not be below 0",
			Field:   "price",
			Code:    render.CodeInvalid,
		}
	}

	if p.PeriodCount.IsZero() {
		return &render.ValidationError{
			Message: "Purchase period is required",
			Field:   "periodCount",
			Code:    render.CodeMissingField,
		}
	}

	if p.Kind == KindRecurring {
		if p.Cycle == enum.CycleNull {
			return &render.ValidationError{
				Message: "Invalid cycle",
				Field:   "cycle",
				Code:    render.CodeInvalid,
			}
		}

		if !p.StartUTC.IsZero() {
			return &render.ValidationError{
				Message: "Recurring price should not set effective time",
				Field:   "startUtc",
				Code:    render.CodeInvalid,
			}
		}

		if !p.EndUTC.IsZero() {
			return &render.ValidationError{
				Message: "Recurring price should not set effective time",
				Field:   "endUtc",
				Code:    render.CodeInvalid,
			}
		}
	}

	if p.Kind == KindOneTime {
		if p.StartUTC.IsZero() {
			return &render.ValidationError{
				Message: "Recurring price should not set effective time",
				Field:   "startUtc",
				Code:    render.CodeMissingField,
			}
		}

		if p.EndUTC.IsZero() {
			return &render.ValidationError{
				Message: "Recurring price should not set effective time",
				Field:   "endUtc",
				Code:    render.CodeMissingField,
			}
		}
	}

	p.Title.String = strings.TrimSpace(p.Title.String)

	ve := validator.New("productId").Required().Validate(p.ProductID)
	if ve != nil {
		return ve
	}

	return p.UpdateParams.Validate()
}

// Price presents the price of a price. It unified prices coming
// from various source, e.g., FTC in-house or Stripe API.
type Price struct {
	ID            string             `json:"id" db:"price_id"`
	Edition                          // Sibling requirement
	Active        bool               `json:"active" db:"is_active"`
	Archived      bool               `json:"archived" db:"archived"` // Once archived, it should never be touched.
	Currency      Currency           `json:"currency" db:"currency"`
	Kind          Kind               `json:"kind" db:"kind"`          // Sibling requirement
	LiveMode      bool               `json:"liveMode" db:"live_mode"` // Sibling requirement
	Nickname      null.String        `json:"nickname" db:"nickname"`
	PeriodCount   ColumnYearMonthDay `json:"periodCount" db:"period_count"`
	ProductID     string             `json:"productId" db:"product_id"` // Sibling requirement. Price's parent.
	StripePriceID string             `json:"stripePriceId" db:"stripe_price_id"`
	Title         null.String        `json:"title" db:"title"`
	UnitAmount    float64            `json:"unitAmount" db:"unit_amount"`
	StartUTC      chrono.Time        `json:"startUtc" db:"start_utc"`
	EndUTC        chrono.Time        `json:"endUtc" db:"end_utc"`
	CreatedUTC    chrono.Time        `json:"createdUtc" db:"created_utc"`
}

func New(p CreationParams, live bool) Price {
	return Price{
		ID:            ids.PriceID(),
		Edition:       p.Edition,
		Active:        false,
		Archived:      false,
		Currency:      "cny",
		Title:         p.Title,
		Kind:          p.Kind,
		LiveMode:      live,
		Nickname:      p.Nickname,
		PeriodCount:   p.PeriodCount,
		ProductID:     p.ProductID,
		StripePriceID: p.StripePriceID,
		UnitAmount:    p.UnitAmount,
		StartUTC:      p.StartUTC,
		EndUTC:        p.EndUTC,
		CreatedUTC:    chrono.TimeNow(),
	}
}

func (p Price) IsZero() bool {
	return p.ID == ""
}

// Update modifies an existing price.
// Only the field listed here is modifiable.
func (p Price) Update(params UpdateParams) Price {
	p.Title = params.Title
	p.Nickname = params.Nickname
	p.StripePriceID = params.StripePriceID
	p.PeriodCount = params.PeriodCount

	return p
}

// Activate put a price on paywall.
func (p Price) Activate() Price {
	p.Active = true

	return p
}

func (p Price) Deactivate() Price {
	p.Active = false

	return p
}

// Archive put a price into archive and no longer usable.
// No idea why I created this.
func (p Price) Archive() Price {
	p.Archived = true
	p.Active = false

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

func (p Price) IsOneTime() bool {
	return p.Kind == KindOneTime
}

func (p Price) IsRecurring() bool {
	return p.Kind == KindRecurring
}
