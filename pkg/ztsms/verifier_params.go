package ztsms

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
)

type VerifierParams struct {
	Mobile      string `json:"mobile"`
	Code        string `json:"code"`
	DeviceToken string `json:"deviceToken"`
}

func (p VerifierParams) ValidateMobile() *render.ValidationError {
	ok := validator.IsMobile(p.Mobile)
	if !ok {
		return &render.ValidationError{
			Message: "Invalid mobile number",
			Field:   "code",
			Code:    render.CodeInvalid,
		}
	}

	return nil
}

func (p VerifierParams) Validate() *render.ValidationError {
	ve := p.ValidateMobile()
	if ve != nil {
		return ve
	}

	return validator.New("code").
		MinLen(6).
		MaxLen(6).
		Required().
		Validate(p.Code)
}
