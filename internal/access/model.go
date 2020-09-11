package access

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/guregu/null"
	"log"
	"net/http"
	"strings"
	"time"
)

var errTokenRequired = errors.New("no access credentials provided")

// GetBearerAuth extracts OAuth access token from request header.
// Authorization: Bearer ***REMOVED***
func GetBearerAuth(req *http.Request) (string, error) {
	authHeader := req.Header.Get("Authorization")
	authForm := req.Form.Get("access_token")

	if authHeader == "" && authForm == "" {
		return "", errTokenRequired
	}

	if authHeader == "" && authForm != "" {
		return authForm, nil
	}

	s := strings.SplitN(authHeader, " ", 2)

	bearerExists := (len(s) == 2) && (strings.ToLower(s[0]) == "bearer")

	log.Printf("Bearer exists: %t", bearerExists)

	if !bearerExists {
		return "", errTokenRequired
	}

	return s[1], nil
}

// OAuthAccess contains the data related to an access token, used
// either by human or machines.
type OAuthAccess struct {
	Token     string      `db:"access_token"`
	Active    bool        `db:"is_active"`
	ExpiresIn null.Int    `db:"expires_in"` // seconds
	CreatedAt chrono.Time `db:"created_utc"`
}

func (o OAuthAccess) Expired() bool {

	if o.ExpiresIn.IsZero() {
		return false
	}

	expireAt := o.CreatedAt.Add(time.Second * time.Duration(o.ExpiresIn.Int64))

	if expireAt.Before(time.Now()) {
		return true
	}

	return false
}
