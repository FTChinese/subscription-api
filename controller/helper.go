package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var logger = log.WithField("project", "subscription-api").
	WithField("package", "controller")

// Param represents a pair of query parameter from URL.
//type Param struct {
//	key   string
//	value string
//}

//// ToBool converts a query parameter to boolean value.
//func (p Param) ToBool() (bool, error) {
//	return strconv.ParseBool(p.value)
//}
//
//// ToString converts a query parameter to string value.
//// Returns error for an empty value.
//func (p Param) ToString() (string, error) {
//	if p.value == "" {
//		return "", fmt.Errorf("%s have empty value", p.key)
//	}
//
//	return p.value, nil
//}
//
//// ToInt converts the value of a query parameter to int64
//func (p Param) ToInt() (int64, error) {
//	if p.value == "" {
//		return 0, fmt.Errorf("%s have empty value", p.key)
//	}
//
//	num, err := strconv.ParseInt(p.value, 10, 0)
//
//	if err != nil {
//		return 0, err
//	}
//
//	return num, nil
//}

// GetURLParam gets a url parameter.
func GetURLParam(req *http.Request, key string) gorest.Param {
	v := chi.URLParam(req, key)

	return gorest.NewParam(key, v)
}
