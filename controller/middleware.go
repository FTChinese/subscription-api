package controller

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/view"
)

const userIDKey = "X-User-Id"
const unionIDKey = "X-Union-Id"

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
func CheckUserID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		userID := req.Header.Get(userIDKey)
		unionID := req.Header.Get(unionIDKey)

		userID = strings.TrimSpace(userID)
		unionID = strings.TrimSpace(unionID)
		if userID == "" && unionID == "" {
			log.WithField("location", "CheckUserID").Info("Missing X-User-Id or X-Union-Id header")

			view.Render(w, view.NewUnauthorized(""))

			return
		}

		req.Header.Set(userIDKey, userID)

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

// CheckUnionID middleware makes sure all request header contains `X-Union-Id` field.
//
// - 401 Unauthorized if request header does not have `X-User-Name`,
// or the value is empty.
func CheckUnionID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		unionID := req.Header.Get(unionIDKey)

		unionID = strings.TrimSpace(unionID)
		if unionID == "" {
			log.WithField("location", "CheckUnionID").Info("Missing X-Unioin-Id header")

			view.Render(w, view.NewUnauthorized(""))

			return
		}

		req.Header.Set(unionIDKey, unionID)

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

// Version show current version of api.
func Version(version, build string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		b := map[string]string{
			"version": version,
			"build":   build,
		}

		view.Render(w, view.NewResponse().NoCache().SetBody(b))
	}
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
