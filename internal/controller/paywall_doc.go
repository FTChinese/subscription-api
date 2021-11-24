package controller

import (
	"database/sql"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"net/http"
)

func (router PaywallRouter) SaveBanner(w http.ResponseWriter, req *http.Request) {
	var params pw.BannerJSON
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(pw.BannerKindDaily); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	banner := params.WithID(pw.BannerKindDaily)

	pwb, err := router.repo.RetrievePaywallDoc(router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if pwb.IsEmpty() {
		pwb = pw.NewPaywallDoc(router.live)
	}

	pwb = pwb.WithBanner(banner)

	id, err := router.repo.CreatePaywallDoc(pwb)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	pwb.ID = id

	_ = render.New(w).OK(pwb)
}

func (router PaywallRouter) SavePromo(w http.ResponseWriter, req *http.Request) {
	var params pw.BannerJSON
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(pw.BannerKindPromo); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	promo := params.WithID(pw.BannerKindPromo)

	pwb, err := router.repo.RetrievePaywallDoc(router.live)
	if err != nil {
		if err != sql.ErrNoRows {
			_ = render.New(w).DBError(err)
			return
		}
	}

	pwb = pwb.WithPromo(promo)

	id, err := router.repo.CreatePaywallDoc(pwb)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	pwb.ID = id

	_ = render.New(w).OK(pwb)
}

func (router PaywallRouter) DropPromo(w http.ResponseWriter, req *http.Request) {
	pwb, err := router.repo.RetrievePaywallDoc(router.live)
	if err != nil {
		if err != sql.ErrNoRows {
			_ = render.New(w).DBError(err)
			return
		}
	}

	pwb = pwb.DropPromo()

	// Save a new version
	id, err := router.repo.CreatePaywallDoc(pwb)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Change id to latest.
	pwb.ID = id

	_ = render.New(w).OK(pwb)
}
