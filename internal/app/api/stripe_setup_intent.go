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

func (router StripeRouter) GetSetupIntent(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	siID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	refresh := xhttp.ParseQueryRefresh(req)
	sugar.Infof("Refreshing setup intent: %t", refresh)

	si, err := router.Env.LoadOrFetchSetupIntent(siID, refresh)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	if si.IsFromStripe {
		go func() {
			err := router.Env.UpsertSetupIntent(si)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(si)
}

func (router StripeRouter) GetSetupPaymentMethod(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	siID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	refresh := xhttp.ParseQueryRefresh(req)

	if !refresh {
		router.loadSetupPaymentMethod(w, siID)
		return
	}

	router.refreshSetupPaymentMethod(w, siID)
}

func (router StripeRouter) loadSetupPaymentMethod(w http.ResponseWriter, setupID string) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	si, err := router.Env.LoadOrFetchSetupIntent(setupID, false)
	if err != nil {
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	if si.IsFromStripe {
		go func() {
			err := router.Env.UpsertSetupIntent(si)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	if si.PaymentMethodID.IsZero() {
		_ = render.New(w).NotFound("Payment method id not set yet")
		return
	}

	pm, err := router.Env.LoadOrFetchPaymentMethod(si.PaymentMethodID.String, false)
	if err != nil {
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

func (router StripeRouter) refreshSetupPaymentMethod(w http.ResponseWriter, setupID string) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	rawSI, err := router.Env.Client.FetchSetupIntent(setupID, true)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleStripeErr(w, err)
	}

	si := stripe.NewSetupIntent(rawSI)
	go func() {
		err := router.Env.UpsertSetupIntent(si)
		if err != nil {
			sugar.Error(err)
		}
	}()

	if rawSI.PaymentMethod == nil || rawSI.PaymentMethod.ID == "" {
		_ = render.New(w).NotFound("Payment method not set")
		return
	}

	pm := stripe.NewPaymentMethod(rawSI.PaymentMethod)

	go func() {
		err := router.Env.UpsertPaymentMethod(pm)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(pm)
}
