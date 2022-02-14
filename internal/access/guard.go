package access

import (
	"database/sql"
	"github.com/FTChinese/go-rest/view"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"log"
	"net/http"
)

type Guard struct {
	env Env
}

func NewGuard(dbs db.ReadWriteMyDBs) Guard {
	return Guard{
		env: NewEnv(dbs),
	}
}

func (g Guard) CheckToken(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			_ = view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		token, err := xhttp.GetAccessToken(req)

		if err != nil {
			log.Printf("Token not found: %s", err)

			_ = view.Render(w, view.NewForbidden("Invalid access token"))
			return
		}

		access, err := g.env.Load(token)

		if err != nil {
			if err == sql.ErrNoRows {
				_ = view.Render(w, view.NewForbidden("Invalid access token"))
				return
			}
			_ = view.Render(w, view.NewDBFailure(err))
			return
		}

		if access.Expired() || !access.Active {
			log.Printf("Token %s is either expired or not active", token)
			_ = view.Render(w, view.NewForbidden("The access token is expired or no longer active"))
			return
		}

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}
