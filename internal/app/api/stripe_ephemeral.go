package api

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// IssueKey creates an ephemeral key.
// https://stripe.com/docs/mobile/android/basic#set-up-ephemeral-key
//
// POST /stripe/customers/{id}/ephemeral-keys?api_version=<version>
// Kept for android app < 6.2.0
func (routes StripeRoutes) IssueKey(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	// Get stripe customer id.
	cusID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	stripeVersion := req.FormValue("api_version")
	if stripeVersion == "" {
		_ = render.New(w).BadRequest("Stripe-Version not found")
		return
	}

	keyData, err := routes.stripeRepo.Client.CreateEphemeralKey(cusID, stripeVersion)
	if err != nil {
		sugar.Error(err)
		err = xhttp.HandleSubsErr(w, err)
		if err == nil {
			return
		}
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	_, err = w.Write(keyData)
	if err != nil {
		sugar.Error(err)
	}
}

func (routes StripeRoutes) SetupWithEphemeral(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	_ = req.ParseForm()

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

	rawSI, rawKey, err := routes.stripeRepo.Client.SetupWithEphemeral(params.Customer)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	paymentSheet := stripe.PaymentSheet{
		ClientSecret:   rawSI.ClientSecret,
		EphemeralKey:   rawKey.Secret,
		CustomerID:     params.Customer,
		PublishableKey: routes.publishableKey,
		LiveMode:       routes.live,
	}

	si := stripe.NewSetupIntent(rawSI)

	go func() {
		err := routes.stripeRepo.UpsertSetupIntent(si)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(paymentSheet)
}
