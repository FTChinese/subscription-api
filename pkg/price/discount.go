package price

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
	"sort"
	"time"
)

// DiscountParams contains fields submitted by client
// when creating a discount.
type DiscountParams struct {
	CreatedBy         string      `json:"createdBy" db:"created_by"`
	Description       null.String `json:"description" db:"discount_desc"`
	Kind              OfferKind   `json:"kind" db:"kind"`
	Percent           null.Int    `json:"percent" db:"percent"`
	dt.DateTimePeriod             // Optional. Zero value indicates permanent discount.
	PriceOff          null.Float  `json:"priceOff" db:"price_off"`
	PriceID           string      `json:"priceId" db:"price_id"`
	Recurring         bool        `json:"recurring" db:"recurring"`
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
	ID string `json:"id" db:"discount_id"`
	DiscountParams
	LiveMode   bool           `json:"liveMode" db:"live_mode"`
	Status     DiscountStatus `json:"status" db:"current_status"`
	CreatedUTC chrono.Time    `json:"createdUtc" db:"created_utc"`
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

	now := time.Now()

	if now.Before(d.StartUTC.Time) || now.After(d.EndUTC.Time) {
		return false
	}

	return true
}

type DiscountListJSON []Discount

func (l DiscountListJSON) Value() (driver.Value, error) {
	if len(l) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (l *DiscountListJSON) Scan(src interface{}) error {
	if src == nil {
		*l = DiscountListJSON{}
		return nil
	}
	switch s := src.(type) {
	case []byte:
		var tmp []Discount
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*l = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to DiscountListJSON")
	}
}

// FindApplicable select an offer from all Offers that a user eligible for.
// Filter criteria:
// * The offer validity period is in effect
// * User is qualified to enjoy
// * Lowest price/Highest discount rate.
// filters - all possible offers a membership currently enjoys, used to narrow down
// offers provided.
// If more than one offer is applicable, use the one with the largest price off.
func (l DiscountListJSON) FindApplicable(filters []OfferKind) Discount {
	// Filter all valid discount offers.
	var filtered = make([]Discount, 0)
	for _, v := range l {
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

func (l DiscountListJSON) FindValid(id string) (Discount, error) {
	for _, v := range l {
		if v.ID == id {
			if v.IsValid() {
				return v, nil
			} else {
				return Discount{}, errors.New("invalid offer selected")
			}
		}
	}

	return Discount{}, errors.New("the requested offer is not found")
}
