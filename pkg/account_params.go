package pkg

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

type PasswordUpdateParams struct {
	FtcID string `json:"-" db:"ftc_id"`
	Old   string `json:"oldPassword"`
	New   string `json:"password" db:"password"` // required. max 128 chars
}

// Validate is only used when a logged-in user changing password.
func (a *PasswordUpdateParams) Validate() *render.ValidationError {
	a.New = strings.TrimSpace(a.New)
	a.Old = strings.TrimSpace(a.Old)

	ve := validator.EnsurePassword(a.New)
	if ve != nil {
		return ve
	}

	return validator.
		New("oldPassword").
		Required().
		Validate(a.Old)
}

type LinkWxParams struct {
	FtcID   string `json:"ftcId"`
	UnionID string `json:"unionId"` // Acquired from header.
}

func (i *LinkWxParams) Validate() *render.ValidationError {
	i.FtcID = strings.TrimSpace(i.FtcID)
	i.UnionID = strings.TrimSpace(i.UnionID)

	return validator.New("ftcId").
		Required().
		Validate(i.FtcID)
}

type UnlinkWxParams struct {
	LinkWxParams
	Anchor enum.AccountKind `json:"anchor"`
}
