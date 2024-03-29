package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"strings"
)

// ProductParams defines the request data to create a new
// product.
type ProductParams struct {
	CreatedBy   string      `json:"createdBy" db:"created_by"` // Only present in creation. Omit it in update.
	Description null.String `json:"description" db:"description"`
	Heading     string      `json:"heading" db:"heading"`
	SmallPrint  null.String `json:"smallPrint" db:"small_print"`
	Tier        enum.Tier   `json:"tier" db:"tier"` // Immutable once a product is created
}

// Validate checks fields to create or update a product.
func (p *ProductParams) Validate(isUpdate bool) *render.ValidationError {
	p.Heading = strings.TrimSpace(p.Heading)
	p.Description.String = strings.TrimSpace(p.Description.String)
	p.SmallPrint.String = strings.TrimSpace(p.SmallPrint.String)

	// Only validate Tier field when creating a product.
	if !isUpdate && p.Tier == enum.TierNull {
		return &render.ValidationError{
			Message: "Tier could not be null",
			Field:   "tier",
			Code:    render.CodeMissingField,
		}
	}

	return validator.
		New("heading").
		Required().
		Validate(p.Heading)
}

type ProductIntroParams struct {
	PriceID string `json:"priceId"`
}

// Product defines a price without plans.
type Product struct {
	ID       string `json:"id" db:"product_id"`
	Active   bool   `json:"active" db:"is_active"` // Indicates whether is product is on paywall
	LiveMode bool   `json:"liveMode" db:"live_mode"`
	ProductParams
	Introductory price.FtcPriceJSON `json:"introductory" db:"introductory"`
	CreatedUTC   chrono.Time        `json:"createdUtc" db:"created_utc"`
	UpdatedUTC   chrono.Time        `json:"updatedUtc" db:"updated_utc"`
}

func NewProduct(params ProductParams, live bool) Product {
	return Product{
		ID:            ids.ProductID(),
		Active:        false,
		LiveMode:      live,
		ProductParams: params,
		Introductory:  price.FtcPriceJSON{},
		CreatedUTC:    chrono.TimeNow(),
		UpdatedUTC:    chrono.Time{},
	}
}

// Update modifies an existing product.
func (p Product) Update(input ProductParams) Product {
	p.Heading = input.Heading
	p.Description = input.Description
	p.SmallPrint = input.SmallPrint
	p.UpdatedUTC = chrono.TimeNow()

	return p
}

func (p Product) Activate() Product {
	p.Active = true
	p.UpdatedUTC = chrono.TimeNow()

	return p
}

// WithIntroPrice set an introductory price on product.
func (p Product) WithIntroPrice(pri price.FtcPrice) Product {
	p.Introductory = price.FtcPriceJSON{
		FtcPrice: pri,
	}

	p.UpdatedUTC = chrono.TimeNow()

	return p
}

// DropIntroPrice removed an introductory price from product.
func (p Product) DropIntroPrice() Product {
	p.Introductory = price.FtcPriceJSON{}
	p.UpdatedUTC = chrono.TimeNow()

	return p
}
