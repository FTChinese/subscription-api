package price

import (
	"fmt"
	"strings"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/conv"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
)

type FtcUpdateParams struct {
	Title         null.String        `json:"title"`
	Nickname      null.String        `json:"nickname"`
	PeriodCount   ColumnYearMonthDay `json:"periodCount"`
	StripePriceID string             `json:"stripePriceId"`
}

// FtcCreationParams is the form data submitted to create a price.
// A new plan is always created under a certain product.
// Therefore, the input data does not have tier field.
type FtcCreationParams struct {
	Edition
	Kind Kind `json:"kind"`
	FtcUpdateParams
	ProductID  string      `json:"productId"`
	StartUTC   chrono.Time `json:"startUtc"`
	EndUTC     chrono.Time `json:"endUtc"`
	UnitAmount float64     `json:"unitAmount"`
}

// Validate checks whether the input data to create a new plan is valid.
// `productTier` is used to specify for which edition of product this plan is created.
// Premium product is not allowed to have a monthly pricing plan.
func (p *FtcCreationParams) Validate() *render.ValidationError {

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

	return validator.New("productId").Required().Validate(p.ProductID)
}

// FtcPrice contains ftc pricing plan.
type FtcPrice struct {
	ID            string             `json:"id" db:"price_id"`
	Edition                          // Sibling requirement
	Active        bool               `json:"active" db:"is_active"`
	Archived      bool               `json:"archived" db:"archived"` // Once archived, it should never be touched and should be hidden from endpoints.
	Currency      Currency           `json:"currency" db:"currency"`
	Kind          Kind               `json:"kind" db:"kind"`          // Sibling requirement
	LiveMode      bool               `json:"liveMode" db:"live_mode"` // Sibling requirement
	Nickname      null.String        `json:"nickname" db:"nickname"`
	PeriodCount   ColumnYearMonthDay `json:"periodCount" db:"period_count"`
	ProductID     string             `json:"productId" db:"product_id"`          // Sibling requirement. Price's parent.
	StripePriceID string             `json:"stripePriceId" db:"stripe_price_id"` // Deprecated, but should not be removed to keep client app compatible.
	Title         null.String        `json:"title" db:"title"`
	UnitAmount    float64            `json:"unitAmount" db:"unit_amount"`
	StartUTC      chrono.Time        `json:"startUtc" db:"start_utc"`
	EndUTC        chrono.Time        `json:"endUtc" db:"end_utc"`
	CreatedUTC    chrono.Time        `json:"createdUtc" db:"created_utc"`
}

func New(p FtcCreationParams, live bool) FtcPrice {
	return FtcPrice{
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

// ActiveID generate a hash for a price active under a product.
// A product should not have duplicate prices with the same features:
// * edition
// * kind
// * mode
func (p FtcPrice) ActiveID() conv.MD5Sum {
	f := p.uniqueFeatures()
	return conv.NewMD5Sum(f)
}

// uniqueFeatures is a string to uniquely
// identify an active price of a product.
func (p FtcPrice) uniqueFeatures() string {
	return fmt.Sprintf("ftc.%s.%s.%s.%s", p.Edition.TierString(), p.Edition.CycleString(), p.Kind, conv.LiveMode(p.LiveMode))
}

func (p FtcPrice) ActiveEntry() ActivePrice {
	return ActivePrice{
		ID:         p.ActiveID().ToHexBin(),
		Source:     PriceSourceFTC,
		ProductID:  p.ProductID,
		PriceID:    p.ID,
		UpdatedUTC: chrono.TimeUTCNow(),
	}
}

func (p FtcPrice) IsZero() bool {
	return p.ID == ""
}

// Update modifies an existing price.
// Only the field listed here is modifiable.
func (p FtcPrice) Update(params FtcUpdateParams) FtcPrice {
	p.Title = params.Title
	p.Nickname = params.Nickname
	p.StripePriceID = params.StripePriceID
	p.PeriodCount = params.PeriodCount

	return p
}

// Activate put a price on paywall.
func (p FtcPrice) Activate() FtcPrice {
	p.Active = true
	p.Archived = false

	return p
}

// Archive put a price into archive and no longer usable.
func (p FtcPrice) Deactivate(archive bool) FtcPrice {
	p.Active = false
	if archive {
		p.Archived = true
	}

	return p
}

func (p FtcPrice) Archive() FtcPrice {
	p.Archived = true
	p.Active = false

	return p
}

// DailyCost calculates the daily average price depending on the cycles.
// Deprecated. Moved to client.
func (p FtcPrice) DailyCost() DailyCost {
	switch p.Cycle {
	case enum.CycleYear:
		return NewDailyCostOfYear(p.UnitAmount)

	case enum.CycleMonth:
		return NewDailyCostOfMonth(p.UnitAmount)
	}

	return DailyCost{}
}

func (p FtcPrice) IsOneTime() bool {
	return p.Kind == KindOneTime
}

func (p FtcPrice) IsRecurring() bool {
	return p.Kind == KindRecurring
}
