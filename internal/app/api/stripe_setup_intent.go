package api

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

func (routes StripeRoutes) CreateSetupIntent(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	var params stripe.CustomerParams
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

	rawSI, err := routes.stripeRepo.Client.CreateSetupIntent(params)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	si := stripe.NewSetupIntent(rawSI)

	go func() {
		err := routes.stripeRepo.UpsertSetupIntent(si)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(si)
}

func (routes StripeRoutes) GetSetupIntent(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	siID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	refresh := xhttp.ParseQueryRefresh(req)
	sugar.Infof("Refreshing setup intent: %t", refresh)

	si, err := routes.stripeRepo.LoadOrFetchSetupIntent(siID, refresh)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	if si.IsFromStripe {
		go func() {
			err := routes.stripeRepo.UpsertSetupIntent(si)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(si)
}

func (routes StripeRoutes) GetSetupPaymentMethod(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	siID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	refresh := xhttp.ParseQueryRefresh(req)

	if !refresh {
		routes.loadSetupPaymentMethod(w, siID)
		return
	}

	routes.refreshSetupPaymentMethod(w, siID)
}

func (routes StripeRoutes) loadSetupPaymentMethod(w http.ResponseWriter, setupID string) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	si, err := routes.stripeRepo.LoadOrFetchSetupIntent(setupID, false)
	if err != nil {
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	if si.IsFromStripe {
		go func() {
			err := routes.stripeRepo.UpsertSetupIntent(si)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	if si.PaymentMethodID.IsZero() {
		_ = render.New(w).NotFound("Payment method id not set yet")
		return
	}

	pm, err := routes.loadPaymentMethod(si.PaymentMethodID.String, false)
	if err != nil {
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	_ = render.New(w).OK(pm)
}

func (routes StripeRoutes) refreshSetupPaymentMethod(w http.ResponseWriter, setupID string) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	rawSI, err := routes.stripeRepo.Client.FetchSetupIntent(setupID, true)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
	}

	si := stripe.NewSetupIntent(rawSI)
	go func() {
		err := routes.stripeRepo.UpsertSetupIntent(si)
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
		err := routes.stripeRepo.UpsertPaymentMethod(pm)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(pm)
}
