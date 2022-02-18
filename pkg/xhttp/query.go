package xhttp

import (
	"net/http"
	"strconv"
)

const (
	KeyRefresh = "refresh"
)

func ParseQueryBool(req *http.Request, key string) bool {
	t, _ := strconv.ParseBool(req.FormValue(key))

	return t
}

func ParseQueryRefresh(req *http.Request) bool {
	return ParseQueryBool(req, KeyRefresh)
}
