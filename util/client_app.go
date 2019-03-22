package util

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/tomasen/realip"
	"net/http"
	"strings"
)

// ClientApp records the header information of a request.
type ClientApp struct {
	ClientType enum.Platform
	Version    null.String
	UserIP     null.String
	UserAgent  null.String
}

// NewClientApp collects information from a request.
func NewClientApp(req *http.Request) ClientApp {
	c := ClientApp{}

	c.ClientType, _ = enum.ParsePlatform(strings.ToLower(req.Header.Get("X-Client-Type")))

	v := strings.TrimSpace(req.Header.Get("X-Client-Version"))
	c.Version = null.NewString(v, v != "")

	// Web app must forward user ip and user agent
	// For other client like Android and iOS, request comes from user's device, not our web app.
	if c.ClientType == enum.PlatformWeb {
		ip := req.Header.Get("X-User-Ip")
		c.UserIP = null.NewString(ip, ip != "")
		ua := req.Header.Get("X-User-Agent")
		c.UserAgent = null.NewString(ua, ua != "")
	} else {
		ip := realip.FromRequest(req)
		c.UserIP = null.NewString(ip, ip != "")
		ua := req.UserAgent()
		c.UserAgent = null.NewString(ua, ua != "")
	}

	return c
}
