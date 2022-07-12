package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
)

// DiscountParams contains fields submitted by client
// when creating a discount.
type DiscountParams struct {
	Description    null.String        `json:"description" db:"discount_desc"`
	Kind           OfferKind          `json:"kind" db:"kind"`
	OverridePeriod ColumnYearMonthDay `json:"overridePeriod" db:"override_period"`
	Percent        null.Int           `json:"percent" db:"percent"`
	PriceOff       null.Float         `json:"priceOff" db:"price_off"`
	PriceID        string             `json:"priceId" db:"price_id"`
	Recurring      bool               `json:"recurring" db:"recurring"`
	dt.TimeSlot                       // Optional. Zero value indicates permanent discount.
	CreatedBy      string             `json:"createdBy" db:"created_by"`
}

func (p DiscountParams) Validate() *render.ValidationError {
	if p.PriceOff.IsZero() || p.PriceOff.Float64 <= 0 {
		return &render.ValidationError{
			Message: "priceOff is required",
			Field:   "priceOff",
			Code:    render.CodeMissingField,
		}
	}

	if !p.StartUTC.IsZero() && !p.EndUTC.IsZero() {
		if p.StartUTC.After(p.EndUTC.Time) {
			return &render.ValidationError{
				Message: "start time must be earlier than end time",
				Field:   "startUtc",
				Code:    render.CodeInvalid,
			}
		}
	}

	return validator.New("description").
		MaxLen(64).
		Validate(p.Description.String)
}

type Discount struct {
	// The id fields started with Disc to avoid conflict when used in ExpandedPlanSchema.
	ID       string         `json:"id" db:"discount_id"`
	LiveMode bool           `json:"liveMode" db:"live_mode"`
	Status   DiscountStatus `json:"status" db:"current_status"`
	DiscountParams
	CreatedUTC chrono.Time `json:"createdUtc" db:"created_utc"`
}

func NewDiscount(params DiscountParams, live bool) Discount {
	return Discount{
		ID:             ids.DiscountID(),
		DiscountParams: params,
		LiveMode:       live,
		Status:         DiscountStatusActive,
		CreatedUTC:     chrono.TimeNow(),
	}
}

func (d Discount) Pause() Discount {
	d.Status = DiscountStatusPaused
	return d
}

func (d Discount) Cancel() Discount {
	d.Status = DiscountStatusCancelled
	return d
}

func (d Discount) IsZero() bool {
	return d.ID == ""
}

func (d Discount) IsValid() bool {
	if d.Status != DiscountStatusActive {
		return false
	}

	if d.PriceOff.IsZero() || d.PriceOff.Float64 <= 0 {
		return false
	}

	if d.StartUTC.IsZero() || d.EndUTC.IsZero() {
		return true
	}

	return d.NowIn()
}
