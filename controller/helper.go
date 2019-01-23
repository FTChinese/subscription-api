package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

var logger = log.WithField("project", "subscription-api").
	WithField("package", "controller")

const (
	msgInvalidURI = "Invalid request URI"
)

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
