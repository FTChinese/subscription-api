package util

import (
	"net/http"

	"github.com/guregu/null"
	log "github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/enum"

	"github.com/tomasen/realip"
)

// ClientApp records the header information of a request.
type ClientApp struct {
	ClientType enum.ClientPlatform
	Version    null.String
	UserIP     null.String
	UserAgent  null.String
}

// NewClientApp collects information from a request.
func NewClientApp(req *http.Request) ClientApp {
	c := ClientApp{}

	c.ClientType = enum.NewPlatform(req.Header.Get("X-Client-Type"))

	v := req.Header.Get("X-Client-Version")
	if v != "" {
		c.Version = null.StringFrom(v)
	}

	// Web app must forward user ip and user agent
	// For other client like Android and iOS, request comes from user's device, not our web app.
	var ip, ua string
	if c.ClientType == enum.PlatformWeb {
		ip = req.Header.Get("X-User-Ip")
		ua = req.Header.Get("X-User-Agent")
	} else {
		ip = realip.FromRequest(req)
		ua = req.UserAgent()
	}

	if ip != "" {
		c.UserIP = null.StringFrom(ip)
	}

	if ua != "" {
		c.UserAgent = null.StringFrom(ua)
	}

	log.WithField("location", "NewClientInfo").Infof("Request IP: %s", ip)

	return c
}
