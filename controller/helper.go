package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"github.com/tomasen/realip"
)

var logger = log.WithField("package", "subscription-api.controller")

const (
	msgInvalidURI  = "Invalid request URI"
	wxNotifyURL    = "http://www.ftacademy.cn/api/v1/callback/wxpay"
	aliNotifyURL   = "http://www.ftacademy.cn/api/v1/callback/alipay"
	aliProductCode = "QUICK_MSECURITY_PAY"
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

// Client represents the essential headers of a request.
type client struct {
	clientType string
	version    string
	userIP     string
	userAgent  string
}

func getClientInfo(req *http.Request) client {
	c := client{}
	c.clientType = req.Header.Get("X-Client-Type")
	if c.clientType == "" {
		c.clientType = "unknown"
	}

	c.version = req.Header.Get("X-Client-Version")

	// Web app must forward user ip and user agent
	// For other client like Android and iOS, request comes from user's device, not our web app.
	if c.clientType == "web" {
		c.userIP = req.Header.Get("X-User-Ip")
		c.userAgent = req.Header.Get("X-User-Agent")
	} else {
		c.userIP = realip.FromRequest(req)
		c.userAgent = req.Header.Get("User-Agent")
	}

	return c
}

type aliAppPayResult struct {
	Response map[string]string `json:"alipay_trade_app_pay_response"`
	Sign     string            `json:"sign"`
	SignType string            `json:"sign_type"`
}

func (r aliAppPayResult) URLValues() url.Values {
	var data url.Values

	for k, v := range r.Response {
		data.Set(k, v)
	}

	data.Set("sign", r.Sign)
	data.Set("sign_type", r.SignType)

	return data
}

// Parse parses input data to struct
func parseJSON(data io.ReadCloser, v interface{}) error {
	dec := json.NewDecoder(data)
	defer data.Close()

	return dec.Decode(v)
}
