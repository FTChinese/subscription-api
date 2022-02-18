package api

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

func (router StripeRouter) CreateSetupIntent(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	var params stripe.SetupIntentParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	rawSI, err := router.Env.Client.CreateSetupIntent(params)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	si := stripe.NewSetupIntent(rawSI)

	go func() {
		err := router.Env.UpsertSetupIntent(si)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(si)
}
