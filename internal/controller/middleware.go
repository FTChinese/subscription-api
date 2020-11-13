package controller

import (
	"fmt"
	"github.com/FTChinese/go-rest/render"
	"net/http"
	"net/http/httputil"
	"strings"

	"log"
)

const (
	userIDKey  = "X-User-Id"
	ftcIDKey   = "X-Ftc-Id"
	unionIDKey = "X-Union-Id"
	appIDKey   = "X-App-Id"
)

// NoCache set Cache-Control request header
func NoCache(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")
		w.Header().Add("Cache-Control", "no-store")
		w.Header().Add("Cache-Control", "must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

// CheckUserID middleware makes sure all request header contains `X-User-Id` field.
//
// - 401 Unauthorized if request header does not have `X-User-Name`,
// or the value is empty.
func UserOrUnionID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		userID := req.Header.Get(userIDKey)
		unionID := req.Header.Get(unionIDKey)

		userID = strings.TrimSpace(userID)
		unionID = strings.TrimSpace(unionID)
		if userID == "" && unionID == "" {
			log.Print("Missing X-User-Id or X-Union-Id header")

			_ = render.New(w).Unauthorized("Missing X-User-Id or X-Union-Id header")

			return
		}

		req.Header.Set(userIDKey, userID)

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

// FtcID middleware makes sure all request header contains `X-User-Id` field.
//
// - 401 Unauthorized if request header does not have `X-User-Name`,
// or the value is empty.
func FtcID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		userID := req.Header.Get(userIDKey)
		ftcID := req.Header.Get(ftcIDKey)

		userID = strings.TrimSpace(userID)
		ftcID = strings.TrimSpace(ftcID)

		if userID == "" && ftcID == "" {
			log.Print("Missing X-Ftc-Id header")

			_ = render.New(w).Unauthorized("")

			return
		}

		req.Header.Set(userIDKey, userID)

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

func GetFtcID(req *http.Request) string {
	ftcID := req.Header.Get(ftcIDKey)
	if ftcID != "" {
		return ftcID
	}

	return req.Header.Get(userIDKey)
}

// CheckUnionID middleware makes sure all request header contains `X-Union-Id` field.
//
// - 401 Unauthorized if request header does not have `X-User-Name`,
// or the value is empty.
func UnionID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		unionID := req.Header.Get(unionIDKey)

		unionID = strings.TrimSpace(unionID)
		if unionID == "" {
			log.Print("Missing X-Union-Id header")

			_ = render.New(w).Unauthorized("Missing X-Union-Id header")

			return
		}

		req.Header.Set(unionIDKey, unionID)

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

func RequireAppID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		appID := req.Header.Get(appIDKey)

		appID = strings.TrimSpace(appID)
		if appID == "" {
			log.Print("Missing X-App-Id header")

			_ = render.New(w).Unauthorized("Missing X-App-Id header")

			return
		}

		req.Header.Set(unionIDKey, appID)

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

// LogRequest print request headers.
func LogRequest(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		dump, err := httputil.DumpRequest(req, false)

		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		}
		log.Printf(string(dump))

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}
