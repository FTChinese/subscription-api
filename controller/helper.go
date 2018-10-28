package controller

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"github.com/smartwalle/alipay/encoding"
	"github.com/tomasen/realip"
)

var logger = log.WithField("package", "subscription-api.controller")

const (
	msgInvalidURI  = "Invalid request URI"
	wxNotifyURL    = "http://www.ftacademy.cn/api/v1/callback/wxpay"
	aliNotifyURL   = "http://www.ftacademy.cn/api/v1/callback/alipay"
	aliProductCode = "QUICK_MSECURITY_PAY"
)

const (
	keySignNodeName = "sign"
	keyAppPayResp   = "alipay_trade_app_pay_response"
	signTypeRSA2    = "RSA2"
	signTypeRSA     = "RSA"
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

type appPayResponse struct {
	Code        string `json:"code"`
	Message     string `json:"msg"`
	AppID       string `json:"app_id"`
	FtcOrderID  string `json:"out_trade_no"`
	AliOrderID  string `json:"trade_no"`
	TotalAmount string `json:"total_amount"`
	DateTime    string `json:"timestamp"`
}

type aliAppPayResult struct {
	Response appPayResponse `json:"alipay_trade_app_pay_response"`
	Sign     string         `json:"sign"`
	SignType string         `json:"sign_type"`
}

// Parse parses input data to struct
func parseJSON(data io.ReadCloser, v interface{}) error {
	dec := json.NewDecoder(data)
	defer data.Close()

	return dec.Decode(v)
}

// This is used to parse Alibaba's stupid trick of using JSON as plain string.
// Copied from github.com/smartwalle/alipay/alipay.go
func extractAppPayResp(rawData string, nodeName string) string {
	var nodeIndex = strings.LastIndex(rawData, nodeName)

	var dataStartIndex = nodeIndex + len(nodeName) + 2
	var signIndex = strings.LastIndex(rawData, "\""+keySignNodeName+"\"")
	var dataEndIndex = signIndex - 1

	var indexLen = dataEndIndex - dataStartIndex
	if indexLen < 0 {
		return ""
	}
	return rawData[dataStartIndex:dataEndIndex]
}

func verifyAliResp(data []byte, sign, singType string, key []byte) (ok bool, err error) {
	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, err
	}

	if singType == signTypeRSA {
		err = encoding.VerifyPKCS1v15(data, signBytes, key, crypto.SHA1)
	} else {
		err = encoding.VerifyPKCS1v15(data, signBytes, key, crypto.SHA256)
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
