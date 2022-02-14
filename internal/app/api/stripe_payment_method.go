package api

import (
	gorest "github.com/FTChinese/go-rest"
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

func (router StripeRouter) ListPaymentMethods(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	ftcID := xhttp.GetFtcID(req.Header)
	p := gorest.GetPagination(req)

	baseAccount, err := router.ReaderRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}
	if baseAccount.StripeID.IsZero() {
		_ = render.New(w).NotFound("Stripe customer not found")
		return
	}

	list, err := router.Env.ListPaymentMethods(baseAccount.StripeID.String, p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}
