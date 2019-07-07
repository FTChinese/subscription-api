package controller

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/FTChinese/go-rest/view"
	log "github.com/sirupsen/logrus"
)

const (
	ftcIDKey   = "X-User-Id"
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
		userID := req.Header.Get(ftcIDKey)
		unionID := req.Header.Get(unionIDKey)

		userID = strings.TrimSpace(userID)
		unionID = strings.TrimSpace(unionID)
		if userID == "" && unionID == "" {
			log.WithField("trace", "CheckUserID").Info("Missing X-User-Id or X-Union-Id header")

			view.Render(w, view.NewUnauthorized("Missing X-User-Id or X-Union-Id header"))

			return
		}

		req.Header.Set(ftcIDKey, userID)

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
		userID := req.Header.Get(ftcIDKey)

		userID = strings.TrimSpace(userID)
		if userID == "" {
			log.WithField("trace", "FtcID").Info("Missing X-User-Id header")

			view.Render(w, view.NewUnauthorized(""))

			return
		}

		req.Header.Set(ftcIDKey, userID)

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
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
			log.WithField("trace", "UnionID").Info("Missing X-Union-Id header")

			view.Render(w, view.NewUnauthorized("Missing X-Union-Id header"))

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
			log.WithField("trace", "RequireAppID").Info("Missing X-App-Id header")

			view.Render(w, view.NewUnauthorized("Missing X-App-Id header"))

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

// DiscountPlans show the current discount plans available.
// func DiscountPlans() func(http.ResponseWriter, *http.Request) {
// 	return func(w http.ResponseWriter, req *http.Request) {
// 		util.Render(w, util.NewResponse().NoCache().SetBody(model.DiscountSchedule))
// 	}
// }

// CurrentPlans see what plan we are using now.
// func () CurrentPlans() func(http.ResponseWriter, *http.Request) {
// 	return func(w http.ResponseWriter, req *http.Request) {
// 		util.Render(
// 			w,
// 			util.
// 				NewResponse().
// 				NoCache().
// 				SetBody(
// 					model.GetCurrentPlans(),
// 				),
// 		)
// 	}
// }
