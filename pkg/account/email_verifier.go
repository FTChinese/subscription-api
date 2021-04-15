package account

import (
	"fmt"
	gorest "github.com/FTChinese/go-rest"
)

// EmailVerifier holds the data used to create an email
// to verify the authenticity of reader's email address.
// For sign-up, the request body contains `email` and `sourceUrl`, and IsSignUp should be true
// For manually requesting verification, request body contains only `sourceUrl`. The email comes from database.
// Client should sent the SourceURL when asking for a verification letter so that
// the client could run under any host.
type EmailVerifier struct {
	Token     string `db:"token"`
	Email     string `db:"email"`
	SourceURL string `db:"source_url"` // The base url to determine verification link.
}

// NewEmailVerifier creates a verification token for an email.
// It sets a default url to build the verification link.
func NewEmailVerifier(email string, sourceURL string) (EmailVerifier, error) {
	token, err := gorest.RandomHex(32)

	if err != nil {
		return EmailVerifier{}, err
	}

	// Provide default url to the verification link
	if sourceURL == "" {
		sourceURL = "https://users.chineseft.com/verify/email"
	}

	return EmailVerifier{
		Token: token,
		Email: email,
		// The default base URL
		SourceURL: sourceURL,
	}, nil
}

// MustNewEmailVerifier creates an EmailVerification instance,
// or panics if error occurred when generating the token.
func MustNewEmailVerifier(email string, sourceURL string) EmailVerifier {
	v, err := NewEmailVerifier(email, sourceURL)
	if err != nil {
		panic(err)
	}

	return v
}

// BuildURL creates a verification link.
func (v EmailVerifier) BuildURL() string {
	return fmt.Sprintf("%s/%s", v.SourceURL, v.Token)
}
