package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

func (router StripeRouter) getPaymentMethod(id string) (stripe.PaymentMethod, error) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	pi, err := router.Env.RetrievePaymentMethod(id)
	if err == nil {
		return pi, nil
	}
	sugar.Error(err)

	rawPM, err := router.Env.Client.FetchPaymentMethod(id)
	if err != nil {
		return stripe.PaymentMethod{}, err
	}

	return stripe.NewPaymentMethod(rawPM), nil
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

	pm, err := router.getPaymentMethod(pmID)
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
