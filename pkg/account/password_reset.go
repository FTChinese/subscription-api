package account

import (
	"fmt"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/guregu/null"
	"time"
)

// PwResetSession hold the to allow resetting password.
// When user is requesting password reset,
// email is sent to server to create the data; an optional
// sourceUrl should be provided by client is user is on
// desktop so that we could create a clickable link in letter.
// If user is using mobile apps, only email is required
// and should generate the AppCode field.
// In both cases the URLToken should be generated.
// The URLToken is directly used to create a clickable link in the email sent to user while mobile apps have to use the AppCode to exchange for the URLToken.
// SourceURL and AppCode are mutually exclusive.
type PwResetSession struct {
	Email     string      `json:"email" db:"email"`
	SourceURL null.String `json:"-" db:"source_url"` // Null for mobile apps
	// A long random string used to build a URL to be used on a web site.
	// It always exists even for mobile devices. To verify a session send from
	// mobiles apps, the request contains Email + AppCode to uniquely identify
	// a row in db since the AppCode is short and duplicate chances are very high.
	// Then the URLToken is sent back so that we could reset password
	// using URLToken, just like in a web page.
	URLToken   string      `json:"token" db:"token"`
	AppCode    null.String `json:"-" db:"app_code"` // A short numeric string sent to reader's email to be used to on mobile devices.
	IsUsed     bool        `json:"-" db:"is_used"`
	ExpiresIn  int64       `json:"-" db:"expires_in"`
	CreatedUTC chrono.Time `json:"-" db:"created_utc"`
}

// NewPwResetSession creates a new PwResetSession instance
// based on request body which contains a required `email`
// field, and an optionally `sourceUrl` field.
func NewPwResetSession(params pkg.ForgotPasswordParams) (PwResetSession, error) {
	token, err := gorest.RandomHex(32)
	if err != nil {
		return PwResetSession{}, err
	}

	if params.SourceURL.IsZero() {
		params.SourceURL = null.StringFrom("https://users.ftchinese.com/password-reset")
	}

	return PwResetSession{
		Email:      params.Email,
		SourceURL:  params.SourceURL,
		URLToken:   token,         // URLToken always exists.
		AppCode:    null.String{}, // Only exists if the request comes from mobile devices.
		IsUsed:     false,
		ExpiresIn:  10800,
		CreatedUTC: chrono.TimeNow(),
	}, nil
}

// MustNewPwResetSession panic on error.
func MustNewPwResetSession(params pkg.ForgotPasswordParams) PwResetSession {
	s, err := NewPwResetSession(params)
	if err != nil {
		panic(err)
	}

	return s
}

func (s PwResetSession) WithSourceURL(url string) PwResetSession {
	if url == "" {
		return s
	}

	s.SourceURL = null.StringFrom(url)

	return s
}

// WithPlatform determines whether the AppCode should be generated.
// For mobile apps, a 6 character string will be generated.
func (s PwResetSession) WithPlatform(p enum.Platform) PwResetSession {

	if p == enum.PlatformIOS || p == enum.PlatformAndroid {
		s.ExpiresIn = 300
		s.AppCode = null.StringFrom(pkg.PwResetCode())
		// For mobile apps we removed the SourceURL
		s.SourceURL = null.String{}
	}

	return s
}

// BuildURL creates password reset link.
// Returns an empty string if AppCode field exists so that
// the template will not render the URL section.
func (s PwResetSession) BuildURL() string {
	if s.AppCode.Valid {
		return ""
	}

	return fmt.Sprintf("%s/%s", s.SourceURL.String, s.URLToken)
}

// IsExpired tests whether an existing PwResetSession is expired.
func (s PwResetSession) IsExpired() bool {
	return s.CreatedUTC.Add(time.Second * time.Duration(s.ExpiresIn)).Before(time.Now())
}
