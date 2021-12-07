package xhttp

import (
	"fmt"
	"github.com/FTChinese/go-rest/render"
	"net/http"
	"net/http/httputil"
	"strings"

	"log"
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

// RequireFtcOrUnionID middleware makes sure all request header contains `X-User-Id` field.
//
// - 401 Unauthorized if request header does not have `X-User-Name`,
// or the value is empty.
func RequireFtcOrUnionID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		userID := GetFtcID(req.Header)
		unionID := GetUnionID(req.Header)

		userID = strings.TrimSpace(userID)
		unionID = strings.TrimSpace(unionID)
		if userID == "" && unionID == "" {
			log.Print("Missing X-User-Id or X-Union-Id header")

			_ = render.New(w).Unauthorized("Missing X-User-Id or X-Union-Id header")

			return
		}

		if userID != "" {
			req.Header.Set(XUserID, userID)
		}

		if unionID != "" {
			req.Header.Set(XUnionID, unionID)
		}

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

func RequireUserIDsQuery(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ftcID := strings.TrimSpace(req.Form.Get("ftc_id"))
		unionID := strings.TrimSpace(req.Form.Get("union_id"))

		if ftcID == "" && unionID == "" {
			_ = render.New(w).Unauthorized("Missing ftc_id or union_id in query parameters")
			return
		}

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

// RequireFtcID middleware makes sure all request header contains `X-User-Id` field.
//
// - 401 Unauthorized if request header does not have `X-User-Name`,
// or the value is empty.
func RequireFtcID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		userID := GetFtcID(req.Header)

		userID = strings.TrimSpace(userID)

		if userID == "" {
			log.Print("Missing X-User-Id header")

			_ = render.New(w).Unauthorized("")

			return
		}

		req.Header.Set(XUserID, userID)

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

// RequireUnionID middleware makes sure all request header contains `X-Union-Id` field.
//
// - 401 Unauthorized if request header does not have `X-User-Name`,
// or the value is empty.
func RequireUnionID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		unionID := GetUnionID(req.Header)

		unionID = strings.TrimSpace(unionID)
		if unionID == "" {
			log.Print("Missing X-Union-Id header")

			_ = render.New(w).Unauthorized("Missing X-Union-Id header")

			return
		}

		req.Header.Set(XUnionID, unionID)

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

func RequireAppID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		appID := GetAppID(req.Header)

		appID = strings.TrimSpace(appID)
		if appID == "" {
			log.Print("Missing X-App-Id header")

			_ = render.New(w).Unauthorized("Missing X-App-Id header")

			return
		}

		req.Header.Set(XAppID, appID)

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

// DumpRequest print request headers.
func DumpRequest(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		dump, err := httputil.DumpRequest(req, false)

		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		}
		log.Printf("\n------Dump request starts------\n%s------Dump request ends------\n", string(dump))

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

func FormParsed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		err := request.ParseForm()
		if err != nil {
			_ = render.New(writer).InternalServerError(err.Error())
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func RequireStaffName(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		staffname := GetStaffName(req.Header)

		staffname = strings.TrimSpace(staffname)
		if staffname == "" {
			log.Print("Missing X-Staff-Name header")

			_ = render.New(w).Unauthorized("Missing X-Staff-Name header")

			return
		}

		req.Header.Set(XStaffName, staffname)

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}
