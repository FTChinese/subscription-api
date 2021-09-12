package input

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/guregu/null"
	"strings"
)

type EmailLoginParams struct {
	Email       string      `json:"email" db:"email"`
	Password    string      `json:"password" db:"password"`
	DeviceToken null.String `json:"deviceToken"` // Required only for android.
}

func (l *EmailLoginParams) Validate() *render.ValidationError {
	l.Email = strings.TrimSpace(l.Email)
	l.Password = strings.TrimSpace(l.Password)

	ve := validator.
		New("email").
		Required().
		MaxLen(64).
		Email().
		Validate(l.Email)

	if ve != nil {
		return ve
	}

	return validator.
		New("password").
		Required().
		MaxLen(64).
		Validate(l.Password)
}

// MobileLinkParams contains the request data when a mobile phone
// is logging in for the first time and is trying to link to an
// existing email account.
type MobileLinkParams struct {
	EmailLoginParams
	Mobile string `json:"mobile"`
}

func (l *MobileLinkParams) Validate() *render.ValidationError {
	ve := l.EmailLoginParams.Validate()
	if ve != nil {
		return ve
	}

	l.Mobile = strings.TrimSpace(l.Mobile)

	return validator.New(l.Mobile).
		Required().
		Validate(l.Mobile)
}

type EmailSignUpParams struct {
	EmailLoginParams
	SourceURL string `json:"sourceUrl"` // Used to compose verification link.
}

func (s *EmailSignUpParams) Validate() *render.ValidationError {
	ve := s.EmailLoginParams.Validate()
	if ve != nil {
		return ve
	}

	return nil
}

type MobileSignUpParams struct {
	EmailSignUpParams
	Mobile string `json:"mobile"`
}

func (s *MobileSignUpParams) Validate() *render.ValidationError {
	ve := s.EmailLoginParams.Validate()
	if ve != nil {
		return ve
	}

	s.Mobile = strings.TrimSpace(s.Mobile)

	return validator.New("mobile").
		Required().
		Validate(s.Mobile)
}

// ForgotPasswordParams is used to create a password reset session.
type ForgotPasswordParams struct {
	Email     string      `json:"email"`
	UseCode   bool        `json:"useCode"`
	SourceURL null.String `json:"sourceUrl"`
}

func (f ForgotPasswordParams) Validate() *render.ValidationError {
	f.Email = strings.TrimSpace(f.Email)

	return validator.EnsureEmail(f.Email)
}

// AppResetPwSessionParams is used to identify a password reset session
// performed on native app.
type AppResetPwSessionParams struct {
	Email   string `schema:"email" json:"email"`
	AppCode string `schema:"code" json:"code"` // hold password reset code for mobile apps.
}

func (i *AppResetPwSessionParams) Validate() *render.ValidationError {
	i.Email = strings.TrimSpace(i.Email)
	i.AppCode = strings.TrimSpace(i.AppCode)

	if ve := validator.EnsureEmail(i.Email); ve != nil {
		return ve
	}

	return validator.
		New("code").
		Required().
		Validate(i.AppCode)
}

// PasswordResetParams contains the data used to reset
type PasswordResetParams struct {
	Token    string `json:"token"`    // identify this session
	Password string `json:"password"` // the new password user submitted
}

func (i *PasswordResetParams) Validate() *render.ValidationError {
	i.Token = strings.TrimSpace(i.Token)
	i.Password = strings.TrimSpace(i.Password)

	ve := validator.
		New("token").
		Required().
		Validate(i.Token)

	if ve != nil {
		return ve
	}

	return validator.EnsurePassword(i.Password)
}
