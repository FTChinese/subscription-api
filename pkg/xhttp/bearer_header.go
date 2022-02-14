package xhttp

import (
	"errors"
	"net/http"
	"strings"
)

var errTokenRequired = errors.New("no access credentials provided")

// ParseBearer extracts Authorization header.
// Authorization: Bearer 19c7d9016b68221cc60f00afca7c498c36c361e3
func ParseBearer(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("empty authorization header")
	}

	s := strings.SplitN(authHeader, " ", 2)

	bearerExists := (len(s) == 2) && (strings.ToLower(s[0]) == "bearer")

	if !bearerExists {
		return "", errors.New("bearer not found")
	}

	return s[1], nil
}

// GetBearerAuth tries to extract the Bearer value from
// Authorization header.
func GetBearerAuth(header http.Header) (string, error) {
	authVal := header.Get("Authorization")

	if authVal == "" {
		return "", errTokenRequired
	}

	return ParseBearer(authVal)
}
