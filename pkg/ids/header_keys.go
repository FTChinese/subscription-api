package ids

import "net/http"

const (
	XUserID  = "X-User-Id"  // Ftc id
	XUnionID = "X-Union-Id" // Wechat union id
	XAppID   = "X-App-Id"   // Wechat app id
)

func GetFtcID(h http.Header) string {
	return h.Get(XUserID)
}

func GetUnionID(h http.Header) string {
	return h.Get(XUnionID)
}

func GetAppID(h http.Header) string {
	return h.Get(XAppID)
}
