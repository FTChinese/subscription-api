package account

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/guregu/null"
)

// Address is a user's physical address
type Address struct {
	FtcID    string      `json:"-" db:"ftc_id"`
	Country  null.String `json:"country" db:"country"`
	Province null.String `json:"province" db:"province"`
	City     null.String `json:"city" db:"city"`
	District null.String `json:"district" db:"district"`
	Street   null.String `json:"street" db:"street"`
	Postcode null.String `json:"postcode" db:"postcode"`
}

// Validate limit input length.
func (a Address) Validate() *render.ValidationError {
	if !a.Country.IsZero() {
		ve := validator.New("country").MaxLen(64).Validate(a.Country.String)
		if ve != nil {
			return ve
		}
	}
	if !a.Province.IsZero() {
		ve := validator.New("province").MaxLen(64).Validate(a.Province.String)
		if ve != nil {
			return ve
		}
	}
	if !a.City.IsZero() {
		ve := validator.New("city").MaxLen(64).Validate(a.City.String)
		if ve != nil {
			return ve
		}
	}
	if !a.District.IsZero() {
		ve := validator.New("district").MaxLen(64).Validate(a.District.String)
		if ve != nil {
			return ve
		}
	}
	if !a.Street.IsZero() {
		ve := validator.New("street").MaxLen(64).Validate(a.Street.String)
		if ve != nil {
			return ve
		}
	}
	if !a.Postcode.IsZero() {
		ve := validator.New("postcode").MaxLen(64).Validate(a.Postcode.String)
		if ve != nil {
			return ve
		}
	}
	return nil
}
