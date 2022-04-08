package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// LoadPaymentMethod load a payment method from db,
// or fetch it from Stripe if not exists.
// If query parameter ?refresh=true is passed,
// it will bypass local db and use Stripe API directly.
func (router StripeRouter) LoadPaymentMethod(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	pmID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	refresh := xhttp.ParseQueryRefresh(req)

	pm, err := router.loadPaymentMethod(pmID, refresh)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	_ = render.New(w).OK(pm)
}

func (router StripeRouter) loadPaymentMethod(pmID string, refresh bool) (stripe.PaymentMethod, error) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Fetch payment method
	pm, err := router.stripeRepo.LoadOrFetchPaymentMethod(pmID, refresh)
	if err != nil {
		return stripe.PaymentMethod{}, err
	}

	// Save it if not save in our db yet.
	if pm.IsFromStripe {
		go func() {
			err := router.stripeRepo.UpsertPaymentMethod(pm)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	return pm, nil
}
