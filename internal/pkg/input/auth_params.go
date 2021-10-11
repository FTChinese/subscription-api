package input

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/guregu/null"
	"strings"
)

type EmailCredentials struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

func (c *EmailCredentials) Validate() *render.ValidationError {
	c.Email = strings.TrimSpace(c.Email)
	c.Password = strings.TrimSpace(c.Password)

	ve := validator.
		New("email").
		Required().
		MaxLen(64).
		Email().
		Validate(c.Email)

	if ve != nil {
		return ve
	}

	return validator.
		New("password").
		Required().
		MaxLen(64).
		Validate(c.Password)
}

type EmailLoginParams struct {
	EmailCredentials
	DeviceToken null.String `json:"deviceToken"` // Required only for android.
}

type EmailSignUpParams struct {
	EmailCredentials
	DeviceToken null.String `json:"deviceToken"` // Required only for android.
	SourceURL   string      `json:"sourceUrl"`   // Used to compose verification link.
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
		Mobile().
		Validate(l.Mobile)
}

// MobileSignUpParams collects signup parameters with mobile.
type MobileSignUpParams struct {
	Mobile      string      `json:"mobile"`
	DeviceToken null.String `json:"deviceToken"` // Required only for android.
}

func (s *MobileSignUpParams) Validate() *render.ValidationError {

	s.Mobile = strings.TrimSpace(s.Mobile)

	return validator.New("mobile").
		Required().
		Mobile().
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
