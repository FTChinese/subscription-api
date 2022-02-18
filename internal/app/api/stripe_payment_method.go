package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// LoadPaymentMethod load a payment method from db,
// or fetch it from Stripe if not exists.
// If query parameter ?refresh=true is passed,
// it will bypass local db and use Stripe API directly.
func (router StripeRouter) LoadPaymentMethod(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	pmID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	refresh := xhttp.ParseQueryRefresh(req)

	pm, err := router.Env.LoadOrFetchPaymentMethod(pmID, refresh)
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
