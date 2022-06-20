package api

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/android"
	"github.com/FTChinese/subscription-api/internal/repository"
	"github.com/FTChinese/subscription-api/internal/repository/apprepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net/http"
)

type AndroidRouter struct {
	dbRepo    apprepo.Env
	cacheRepo repository.CacheRepo
	logger    *zap.Logger
}

func NewAndroidRouter(dbs db.ReadWriteMyDBs, c *cache.Cache, logger *zap.Logger) AndroidRouter {
	return AndroidRouter{
		dbRepo:    apprepo.New(dbs),
		cacheRepo: repository.NewCacheRepo(c),
		logger:    logger,
	}
}

// CreateRelease inserts the metadata for a new Android release.
//
// POST /android/releases
//
// Body: {versionName: string, versionCode: int, body?: string, apkUrl: string}
func (router AndroidRouter) CreateRelease(w http.ResponseWriter, req *http.Request) {
	var input android.ReleaseInput

	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := input.ValidateCreation(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	release, err := router.dbRepo.CreateRelease(android.NewRelease(input))
	if err != nil {
		if db.IsAlreadyExists(err) {
			_ = render.New(w).Unprocessable(render.NewVEAlreadyExists("versionName"))
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(release)
}

// UpdateRelease updates a single release.
//
// PATCH /android/releases/{versionName}
//
// Body {body: string, apkUrl: string}
func (router AndroidRouter) UpdateRelease(w http.ResponseWriter, req *http.Request) {
	versionName, _ := xhttp.GetURLParam(req, "versionName").ToString()

	var input android.ReleaseInput
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	input.VersionName = versionName

	if ve := input.ValidateUpdate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	current, err := router.dbRepo.RetrieveRelease(versionName)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	updated := current.Update(input)

	err = router.dbRepo.UpdateRelease(updated)
	if err != nil {
		if db.IsAlreadyExists(err) {
			_ = render.New(w).Unprocessable(render.NewVEAlreadyExists("versionCode"))
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(updated)
}

// DeleteRelease deletes a single release
//
// DELETE /android/releases/:versionName
func (router AndroidRouter) DeleteRelease(w http.ResponseWriter, req *http.Request) {
	versionName, _ := xhttp.GetURLParam(req, "versionName").ToString()

	if err := router.dbRepo.DeleteRelease(versionName); err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).NoContent()
}

func (router AndroidRouter) loadLatestRelease(refresh bool) (android.Release, error) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	if !refresh {
		sugar.Infof("Loading android latest release from cache")
		release, err := router.cacheRepo.LoadAndroidLatest()
		if err == nil {
			return release, nil
		}
		sugar.Error(err)
	}

	sugar.Infof("Loading android latest release from db")
	release, err := router.dbRepo.RetrieveLatest()
	if err != nil {
		return android.Release{}, err
	}

	// Cache it.
	router.cacheRepo.AndroidLatest(release)

	return release, nil
}

// LatestRelease show the latest release of the Android app.
// Query parameters:
// - refresh=<true|false>, bust cache.
// The latest release is cached if retrieved from db.
// Use `refresh` to circumvent cache. Usually you do not need to
// do this since the cache expires in 2 hours.
func (router AndroidRouter) LatestRelease(w http.ResponseWriter, req *http.Request) {
	refresh := xhttp.ParseQueryRefresh(req)

	r, err := router.loadLatestRelease(refresh)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(r)
}

// ListReleases retrieves all releases by sorting version code
// in descending order.
//
// GET /android/releases?page=<number>&per_page=<number>
func (router AndroidRouter) ListReleases(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()

	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	pagination := gorest.GetPagination(req)

	releases, err := router.dbRepo.ListReleases(pagination)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(releases)
}

// LoadOneRelease retrieves a release by version name
//
// GET /android/releases/{versionName}
//
// Warning: DO NOT remove this method! In release v0.7 it is removed because I
// forgot the Android app has a current release page to show the log of current version.
func (router AndroidRouter) LoadOneRelease(w http.ResponseWriter, req *http.Request) {
	versionName, err := xhttp.GetURLParam(req, "versionName").ToString()

	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	release, err := router.dbRepo.RetrieveRelease(versionName)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(release)
}
