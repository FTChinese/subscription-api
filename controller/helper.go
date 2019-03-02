package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

var logger = log.WithField("project", "subscription-api").
	WithField("package", "controller")

const (
	msgInvalidURI = "Invalid request URI"
)

// Param represents a pair of query parameter from URL.
type Param struct {
	key   string
	value string
}

// ToBool converts a query parameter to boolean value.
func (p Param) ToBool() (bool, error) {
	return strconv.ParseBool(p.value)
}

// ToString converts a query parameter to string value.
// Returns error for an empty value.
func (p Param) ToString() (string, error) {
	if p.value == "" {
		return "", fmt.Errorf("%s have empty value", p.key)
	}

	return p.value, nil
}

// ToInt converts the value of a query parameter to int64
func (p Param) ToInt() (int64, error) {
	if p.value == "" {
		return 0, fmt.Errorf("%s have empty value", p.key)
	}

	num, err := strconv.ParseInt(p.value, 10, 0)

	if err != nil {
		return 0, err
	}

	return num, nil
}

// GetQueryParam gets a pair of query parameter from URL.
func GetQueryParam(req *http.Request, key string) Param {
	v := req.Form.Get(key)

	return Param{
		key:   key,
		value: strings.TrimSpace(v),
	}
}

// GetURLParam gets a url parameter.
func GetURLParam(req *http.Request, key string) Param {
	v := chi.URLParam(req, key)

	return Param{
		key:   key,
		value: strings.TrimSpace(v),
	}
}

type paramValue string

func (v paramValue) isEmpty() bool {
	return string(v) == ""
}

func (v paramValue) toInt() (int64, error) {
	if v.isEmpty() {
		return 0, errors.New("query: empty value")
	}

	num, err := strconv.ParseInt(string(v), 10, 0)

	if err != nil {
		return 0, err
	}

	return num, nil
}

// Convert paramValue to boolean value.
// Returns error if the paramValue cannot be converted.
func (v paramValue) toBool() (bool, error) {
	return strconv.ParseBool(string(v))
}

func (v paramValue) toString() string {
	return string(v)
}

func getQueryParam(req *http.Request, key string) paramValue {
	value := req.Form.Get(key)

	return paramValue(value)
}

func getURLParam(req *http.Request, key string) paramValue {
	value := chi.URLParam(req, key)

	return paramValue(value)
}

// Parse parses input data to struct
func parseJSON(data io.ReadCloser, v interface{}) error {
	dec := json.NewDecoder(data)
	defer data.Close()

	return dec.Decode(v)
}
