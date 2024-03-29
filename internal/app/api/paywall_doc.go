package api

import (
	"database/sql"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"net/http"
)

// SaveBanner saves a new banner. It actually creates a new
// row based on latest paywall.
// Request body:
// - heading: string;
// - subHeading?: string;
// - coverUrl?: string;
// - content?: string;
// - terms?: string;
func (router PaywallRouter) SaveBanner(w http.ResponseWriter, req *http.Request) {
	var params reader.BannerJSON
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(reader.BannerKindDaily); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	banner := params.WithID(reader.BannerKindDaily)

	pwb, err := router.paywallRepo.RetrievePaywallDoc(router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if pwb.IsEmpty() {
		pwb = reader.NewPaywallDoc(router.live)
	}

	pwb = pwb.WithBanner(banner)

	id, err := router.productRepo.CreatePaywallDoc(pwb)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	pwb.ID = id

	_ = render.New(w).OK(pwb)
}

// SavePromo attaches a promotion banner to paywall.
// Request body:
// - heading: string;
// - subHeading?: string;
// - coverUrl?: string;
// - content?: string;
// - terms?: string;
// - startUtc: string;
// - endUtc: string;
func (router PaywallRouter) SavePromo(w http.ResponseWriter, req *http.Request) {
	var params reader.BannerJSON
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(reader.BannerKindPromo); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	promo := params.WithID(reader.BannerKindPromo)

	pwb, err := router.paywallRepo.RetrievePaywallDoc(router.live)
	if err != nil {
		if err != sql.ErrNoRows {
			_ = render.New(w).DBError(err)
			return
		}
	}

	pwb = pwb.WithPromo(promo)

	id, err := router.productRepo.CreatePaywallDoc(pwb)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	pwb.ID = id

	_ = render.New(w).OK(pwb)
}

func (router PaywallRouter) DropPromo(w http.ResponseWriter, req *http.Request) {
	pwb, err := router.paywallRepo.RetrievePaywallDoc(router.live)
	if err != nil {
		if err != sql.ErrNoRows {
			_ = render.New(w).DBError(err)
			return
		}
	}

	pwb = pwb.DropPromo()

	// Save a new version
	id, err := router.productRepo.CreatePaywallDoc(pwb)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Change id to latest.
	pwb.ID = id

	_ = render.New(w).OK(pwb)
}
