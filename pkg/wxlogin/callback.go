package wxlogin

import (
	"fmt"
	"net/url"
)

const (
	CallbackAppNextUser CallbackApp = iota
	CallbackAppFtaReader
)

type CallbackApp int

var cbAppNames = []string{
	"next-user",
	"fta-reader",
}

func (a CallbackApp) String() string {
	if a < CallbackAppNextUser || a > CallbackAppFtaReader {
		return ""
	}

	return cbAppNames[a]
}

var callbackURLs = map[CallbackApp]url.URL{
	// the callback url of next-user
	// to which this API should forward OAuth's code and state.
	CallbackAppNextUser: {
		Scheme:   "https",
		Host:     "users.ftchinese.com",
		Path:     "/login/wechat/callback",
		RawQuery: "",
	},
	CallbackAppFtaReader: {
		Scheme:   "https",
		Opaque:   "next.ftacademy.cn",
		Path:     "/reader/oauth/callback",
		RawQuery: "",
	},
}

// GetCallbackURL finds the callback url for an app, and appends query parameter to it.
func GetCallbackURL(app CallbackApp, query url.Values) (string, error) {
	u, ok := callbackURLs[app]
	if !ok {
		return "", fmt.Errorf("url to redirect to %s not found", app)
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}
