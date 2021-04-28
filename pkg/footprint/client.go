package footprint

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/tomasen/realip"
	"net/http"
	"strings"
)

// Client records the header information of a request:
// * `X-Client-Type: web | ios | android`
// * `X-Client-Version: 1.2.1`
// * `X-User-Ip: 1.2.3.4`
// * `X-User-Agent: chrome`, only applicable to web app which forwards user agent here.
// * `User-Agent: okhttp` only applicable to mobile devices.
type Client struct {
	Platform  enum.Platform `db:"platform"`
	Version   null.String   `db:"client_version"`
	UserIP    null.String   `db:"user_ip"`
	UserAgent null.String   `db:"user_agent"`
}

// NewClient collects information from a request.
func NewClient(req *http.Request) Client {
	c := Client{}

	c.Platform, _ = enum.ParsePlatform(strings.ToLower(req.Header.Get("X-Client-Type")))

	v := strings.TrimSpace(req.Header.Get("X-Client-Version"))
	c.Version = null.NewString(v, v != "")

	// Web app must forward user ip and user agent
	// For other client like Android and iOS, request comes from user's device, not our web app.
	if c.Platform == enum.PlatformWeb {
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
