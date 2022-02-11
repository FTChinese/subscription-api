package api

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/apprepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

type AppRouter struct {
	repo apprepo.Env
}

func NewAppRouter(dbs db.ReadWriteMyDBs) AppRouter {
	return AppRouter{
		repo: apprepo.New(dbs),
	}
}

// AndroidLatest show the latest release of the Android app.
func (router AppRouter) AndroidLatest(w http.ResponseWriter, req *http.Request) {
	r, err := router.repo.Latest()
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(r)
}

// AndroidList retrieves all releases by sorting version code
// in descending order.
//
// GET /android/releases?page=<number>&per_page=<number>
func (router AppRouter) AndroidList(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()

	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	pagination := gorest.GetPagination(req)

	releases, err := router.repo.Releases(pagination)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(releases)
}

// AndroidSingle retrieves a release by version name
//
// GET /android/releases/{versionName}
//
// Warning: DO NOT remove this method! In release v0.7 it is removed because I
// forgot the Android app has a current release page to show the log of current version.
func (router AppRouter) AndroidSingle(w http.ResponseWriter, req *http.Request) {
	versionName, err := xhttp.GetURLParam(req, "versionName").ToString()

	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	release, err := router.repo.SingleRelease(versionName)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(release)
}
