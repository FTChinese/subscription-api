package xhttp

import (
	"net/http"
	"strconv"
)

const (
	KeyRefresh = "refresh"
)

func ParseQueryBool(req *http.Request, key string) bool {
	v := req.FormValue(key)
	if v == "" {
		return false
	}

	// Ignore error so that always returns false in case of parsing failure.
	t, _ := strconv.ParseBool(v)

	return t
}

func ParseQueryRefresh(req *http.Request) bool {
	return ParseQueryBool(req, KeyRefresh)
}
