package util

import (
	"net/http"

	"gitlab.com/ftchinese/subscription-api/enum"

	"github.com/tomasen/realip"
)

// RequestClient contains essential data of a user who are placing an order
// It is exported since both controller and model need it.
type RequestClient struct {
	ClientType enum.ClientPlatform
	Version    string
	UserIP     string
	UserAgent  string
}

// GetClient extracts client data from a request.
func GetClient(req *http.Request) RequestClient {
	c := RequestClient{}

	c.ClientType = enum.NewPlatform(req.Header.Get("X-Client-Type"))

	c.Version = req.Header.Get("X-Client-Version")

	// Web app must forward user ip and user agent
	// For other client like Android and iOS, request comes from user's device, not our web app.
	if c.ClientType == enum.PlatformWeb {
		c.UserIP = req.Header.Get("X-User-Ip")
		c.UserAgent = req.Header.Get("X-User-Agent")
	} else {
		c.UserIP = realip.FromRequest(req)
		c.UserAgent = req.UserAgent()
	}

	return c
}
