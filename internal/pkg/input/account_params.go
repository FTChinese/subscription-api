package input

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/guregu/null"
	"strings"
)

type EmailUpdateParams struct {
	Email     string      `json:"email"`
	SourceURL null.String `json:"sourceUrl"`
}

func (i *EmailUpdateParams) Validate() *render.ValidationError {
	i.Email = strings.TrimSpace(i.Email)
	url := strings.TrimSpace(i.SourceURL.String)

	i.SourceURL = null.NewString(url, url != "")

	return validator.EnsureEmail(i.Email)
}

type ReqEmailVrfParams struct {
	SourceURL null.String `json:"sourceUrl"`
}

type NameUpdateParams struct {
	UserName string `json:"userName"`
}

func (p *NameUpdateParams) Validate() *render.ValidationError {
	p.UserName = strings.TrimSpace(p.UserName)

	return validator.New("userName").
		Required().
		MaxLen(64).
		Validate(p.UserName)
}

// PasswordUpdateParams is used to hold data to change password.
// Use CurrentPassword to verify before changing to NewPassword.
type PasswordUpdateParams struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
	Old             string `json:"oldPassword"` // Deprecated. Use CurrentPassword.
	New             string `json:"password"`    // Deprecated. Use NewPassword.
}

// Validate is only used when a logged-in user changing password.
func (a *PasswordUpdateParams) Validate() *render.ValidationError {
	a.New = strings.TrimSpace(a.New)
	a.Old = strings.TrimSpace(a.Old)
	a.CurrentPassword = strings.TrimSpace(a.CurrentPassword)
	a.NewPassword = strings.TrimSpace(a.NewPassword)

	// Validate legacy fields.
	if a.New != "" || a.Old != "" {
		ve := validator.EnsurePassword(a.New)
		if ve != nil {
			return ve
		}

		ve = validator.
			New("oldPassword").
			Required().
			Validate(a.Old)

		if ve != nil {
			return ve
		}

		a.CurrentPassword = a.Old
		a.NewPassword = a.New

		return nil
	}

	ve := validator.
		New("currentPassword").
		Required().
		MaxLen(64).
		MinLen(8).
		Validate(a.CurrentPassword)

	if ve != nil {
		return ve
	}

	return validator.
		New("newPassword").
		Required().
		MaxLen(64).
		MinLen(8).
		Validate(a.NewPassword)
}

type LinkWxParams struct {
	FtcID   string `json:"ftcId"`
	UnionID string `json:"unionId"` // Acquired from header.
}

func (i *LinkWxParams) Validate() *render.ValidationError {
	i.FtcID = strings.TrimSpace(i.FtcID)
	i.UnionID = strings.TrimSpace(i.UnionID)

	ve := validator.New("ftcId").
		Required().
		Validate(i.FtcID)

	if ve != nil {
		return ve
	}

	return validator.New("unionId").
		Required().
		Validate(i.UnionID)
}

type UnlinkWxParams struct {
	LinkWxParams
	Anchor enum.AccountKind `json:"anchor"`
}
