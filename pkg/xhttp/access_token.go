package xhttp

import (
	"net/http"
	"strings"
)

func GetTokenFromQuery(req *http.Request) (string, error) {
	token := req.Form.Get("access_token")
	if strings.TrimSpace(token) == "" {
		return "", errTokenRequired
	}

	return token, nil
}

// GetAccessToken extract oauth token from http request.
// It first tries to find it from `Authorization: Bearer xxxxx`
// header, then fallback to url query parameter `access_token`
// field.
// If nothing is found, returns error.
func GetAccessToken(req *http.Request) (string, error) {
	authHeader, err := GetBearerAuth(req.Header)
	if err == nil {
		return authHeader, nil
	}

	return GetTokenFromQuery(req)
}
