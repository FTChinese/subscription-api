package util

import (
	"net/http"

	"github.com/tomasen/realip"
)

// RequestClient contains essential data of a user who are placing an order
type RequestClient struct {
	ClientType string
	Version    string
	UserIP     string
	UserAgent  string
}

// NewRequestClient extracts client data from a request.
func NewRequestClient(req *http.Request) RequestClient {
	c := RequestClient{}

	c.ClientType = req.Header.Get("X-Client-Type")
	if c.ClientType == "" {
		c.ClientType = "unknown"
	}

	c.Version = req.Header.Get("X-Client-Version")

	// Web app must forward user ip and user agent
	// For other client like Android and iOS, request comes from user's device, not our web app.
	if c.ClientType == "web" {
		c.UserIP = req.Header.Get("X-User-Ip")
		c.UserAgent = req.Header.Get("X-User-Agent")
	} else {
		c.UserIP = realip.FromRequest(req)
		c.UserAgent = req.Header.Get("User-Agent")
	}

	return c
}
