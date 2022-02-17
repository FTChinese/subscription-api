package api

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// UpsertPaymentMethod insert or update a new payment method
// to our db as requested by client.
// Used when a new payment method is created on the client
// but the server on our side has no idea it exists.
func (router StripeRouter) UpsertPaymentMethod(w http.ResponseWriter, req *http.Request) {
	var pmIDParam input.IDParam
	if err := gorest.ParseJSON(req.Body, &pmIDParam); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := pmIDParam.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	rawPM, err := router.Env.Client.FetchPaymentMethod(pmIDParam.ID)
	if err != nil {
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	pm := stripe.NewPaymentMethod(rawPM)
	err = router.Env.UpsertPaymentMethod(pm)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(pm)
}

func (router StripeRouter) LoadPaymentMethod(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	pmID, err := xhttp.GetURLParam(req, "id").ToString()

	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	pm, err := router.Env.LoadOrFetchPaymentMethod(pmID)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	if pm.IsFromStripe {
		go func() {
			err := router.Env.UpsertPaymentMethod(pm)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(pm)
}
