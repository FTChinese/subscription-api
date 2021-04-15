package account

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/guregu/null"
	"strings"
)

type BaseProfile struct {
	FtcID      string      `json:"id" db:"ftc_id"`
	Gender     enum.Gender `json:"gender" db:"gender"`          // Optional
	FamilyName null.String `json:"familyName" db:"family_name"` // Optional. Max 50 chars.
	GivenName  null.String `json:"givenName" db:"given_name"`   // Optional. Max 50 chars
	Birthday   null.String `json:"birthday" db:"birthday"`      // Optional. Max 10 chars.
	CreatedAt  chrono.Time `json:"createdAt" db:"created_utc"`
	UpdatedAt  chrono.Time `json:"updatedAt" db:"updated_utc"`
}

// Validate performs validation when user update profile: gender, familyName, givenName, birthdate.
// Other fields are updated separately.
func (p *BaseProfile) Validate() *render.ValidationError {
	if !p.FamilyName.IsZero() {
		p.FamilyName.String = strings.TrimSpace(p.FamilyName.String)
	}
	if !p.GivenName.IsZero() {
		p.GivenName.String = strings.TrimSpace(p.GivenName.String)
	}

	if !p.Birthday.IsZero() {
		p.Birthday.String = strings.TrimSpace(p.Birthday.String)
	}

	if !p.FamilyName.IsZero() {
		ve := validator.New("familyName").MaxLen(64).Validate(p.FamilyName.String)
		if ve != nil {
			return ve
		}
	}

	if !p.GivenName.IsZero() {
		ve := validator.New("givenName").MaxLen(64).Validate(p.GivenName.String)
		if ve != nil {
			return ve
		}
	}

	if !p.Birthday.IsZero() {
		return validator.New("birthday").MaxLen(64).Validate(p.Birthday.String)
	}

	return nil
}

// Profile contains all data of a user.
// This is used as API output.
type Profile struct {
	BaseProfile
	Address Address `json:"address"`
}

// ProfileSchema is used as db scan target.
type ProfileSchema struct {
	BaseProfile
	Address
}

func (s ProfileSchema) Profile() Profile {
	return Profile{
		BaseProfile: s.BaseProfile,
		Address:     s.Address,
	}
}
